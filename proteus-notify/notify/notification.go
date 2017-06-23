package notify

import (
	"time"
	"errors"

	"github.com/spf13/viper"
	"path/filepath"

	"github.com/NaySoftware/go-fcm"
	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

type PushNotification struct {
	Tokens		[]string
	Platform	string
	Priority	string
	Data		map[string]interface{}
	Retry		int
	Topic		string

	// iOS specific
	Expiration	time.Time
	ApnsID		string

	// Android specific
	TimeToLive	int
	DryRun		bool
}

func InitApnsClient() error {
	// In development environment we don't actually speak to APNs
	if viper.GetString("core.environment") == "development" {
		return nil
	}

	var (
		err error
		apnKeyPath = viper.GetString("apn.key-path")
		apnKeyPassword = viper.GetString("apn.key-password")
		isProduction = viper.GetBool("apn.production")
	)
	ctx.Debugf("Using key path: %s", apnKeyPath)
	ext := filepath.Ext(apnKeyPath)
	switch ext {
	case ".p12":
		CertificatePemIos, err = certificate.FromP12File(apnKeyPath, apnKeyPassword)
	case ".pem":
		CertificatePemIos, err = certificate.FromPemFile(apnKeyPath, apnKeyPassword)
	default:
		err = errors.New("wrong certificate key extension")
	}
	if err != nil {
		ctx.WithError(err).Errorf("certificate error (ext: %s)", ext)
		return err
	}

	if isProduction {
		ApnsClient = apns.NewClient(CertificatePemIos).Production()
		return nil
	}
	ApnsClient = apns.NewClient(CertificatePemIos).Development()
	return nil
}

func InitWorkers(workerNum int, queueSize int) {
	ctx.Debugf("worker number: %d, queue size: %d", workerNum, queueSize)
	QueueNotification = make(chan PushNotification, queueSize)
	for i := 0; i < workerNum; i++ {
		go startWorker()
	}
}

func startWorker() {
	for {
		notification := <-QueueNotification
		switch notification.Platform {
		case "ios":
			PushToApn(notification)
		case "android":
			PushToFcm(notification)
		default:
			ctx.Errorf("unsupported platform %s", notification.Platform)
		}
	}
}

func PushToAny(req PushNotification) {
	QueueNotification <- req
}

func PushToApn(req PushNotification) {
	ctx.Debug("Pushing iOS notification to APN")
	
	var retryCount = 0
	var retryAfter = 1
	var maxRetry = viper.GetInt("apn.max-retry")

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}
	
	var isDone = false

	for isDone == false {
		var toRetryTokens []string

		notification := MakeApnNotification(req)

		for _, token := range req.Tokens {
			ctx.Debugf("Sending APN notification to token \"%s\"", token)
			if viper.GetString("core.environment") == "development" {
				ctx.Infof("I would have sent a %s notification with token %s", req.Platform, token)
				continue
			}

			notification.DeviceToken = token
			res, err := ApnsClient.Push(notification)

			if err != nil {
				// Maybe write this somewhere?
				ctx.WithError(err).Error("failed to notify ios")
				toRetryTokens = append(toRetryTokens, token)
				continue
			}

			if res.Sent() {
				ctx.Debugf("sent %v", res.ApnsID)
			} else {
				// res.Reason are defined:
				// https://github.com/sideshow/apns2/blob/master/response.go#L14-L105
				// XXX when we see `ReasonExpiredProviderToken` and
				// `ReasonBadDeviceToken` we probably want to write this
				// somewhere and stop trying to send messages to these
				// devices.
				ctx.Errorf("failed to send %v %v %v",
									res.StatusCode,
									res.ApnsID,
									res.Reason)
				toRetryTokens = append(toRetryTokens, token)
				continue
			}
		}
		if len(toRetryTokens) == 0 || retryCount >= maxRetry {
			isDone = true
		} else {
			time.Sleep(time.Duration(retryAfter) * time.Second)
			retryAfter = retryAfter*2
			retryCount++
			req.Tokens = toRetryTokens
		}
	}
}


func MakeApnNotification(req PushNotification) *apns.Notification {
	notification := &apns.Notification{
		ApnsID: req.ApnsID,
		Topic:  req.Topic,
	}

	if len(req.Priority) > 0 && req.Priority == "normal" {
		notification.Priority = apns.PriorityLow
	}

	payload := payload.NewPayload()

	for k, v := range req.Data {
		payload.Custom(k, v)
	}

	notification.Payload = payload

	return notification
}

func PushToFcm(req PushNotification) {
	ctx.Debug("Pushing Android notification to FCM")
	
	var retryCount = 0
	var retryAfter = 1
	var maxRetry = viper.GetInt("fcm.max-retry")

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}
	
	var isDone = false

	for isDone == false {
		var toRetryTokens []string

		notification := MakeFcmNotification(req)
		ctx.Debugf("Sending a FCM notification")
		if viper.GetString("core.environment") == "development" {
			ctx.Infof("I would have sent a %s notification with token %v", req.Platform, req.Tokens)
			isDone = true
			break
		}

		res, err := notification.Send()
		if err == nil {
			ctx.Debugf("SENT a FCM notification")
			isDone = true
			break
		} else {
			ctx.WithError(err).Error("failed to notify")
			for k, result := range res.Results {
				if result["Error"] != "" {
					toRetryTokens = append(toRetryTokens, req.Tokens[k])
				}
			}
		}
		if len(toRetryTokens) == 0 || retryCount >= maxRetry {
			isDone = true
		} else {
			time.Sleep(time.Duration(retryAfter) * time.Second)
			retryAfter = retryAfter*2
			retryCount++
			req.Tokens = toRetryTokens
		}
	}
}

func MakeFcmNotification(req PushNotification) *fcm.FcmClient {
	// XXX the FCM API only supports up to 1k devices. We should have checks
	// for that here and split them up into < 1000 chunks
	// https://firebase.google.com/docs/cloud-messaging/http-server-ref
	// 1000 devices

	notification := fcm.NewFcmClient(viper.GetString("fcm.server-key"))
	notification.NewFcmRegIdsMsg(req.Tokens, req.Data)
	notification.Message.To = req.Topic

	if len(req.Priority) > 0 {
		notification.SetPriority(req.Priority)
	}

	return notification
}
