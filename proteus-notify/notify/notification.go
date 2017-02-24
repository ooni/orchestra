package notify

import (
	"github.com/spf13/viper"
	"path/filepath"
	apns "github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

type PushNotification struct {
	Tokens		[]string
	Platform	string
	Priority	string
	Data		interface{}

	// iOS specific
	Expiration	time.Time
	ApnsID		string
	Topic		string

	// Android specific
	ServerKey	string
	TimeToLive	int
	DryRun		bool
	Condition	string
}

func InitAPNSClient() error {
	var (
		err error
		apnKeyPath = viper.GetString("apn-key-path")
		apnKeyPassword = viper.GetString("apn-key-password")
		isProduction = viper.GetBool("apn-production")
	)
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
		ctx.WithError(err).Error("certificate error")
		return err
	}

	if isProduction {
		ApnsClient = apns.NewClient(CertificatePemIos).Production()
		return
	}
	ApnsClient = apns.NewClient(CertificatePemIos).Development()
	return
}

func InitWorkers(workerNum int64, queueSize int64) {
	ctx.Debugf("worker number: %d, queue size: %d", workerNum, queueSize)
	QueueNotification = make(chan PushNotification, queueNum)
	for i := int64(0); i < workerNum; i++ {
		go startWorker()
	}
}

func startWorker() {
	for {
		notification := <-QueueNotification
		switch notification.Platform {
		case "ios":
			PushToAPN(notification)
		case "android":
			PushToFCM(notification)
		default:
			ctx.Errorf("Unsupported platform %s", notification.Platform)
		}
	}
}


func PushToAPN(req PushNotification) {
	ctx.Debug("Pushing iOS notification to APN")
	notification := MakeAPNNotification(req)
}


func MakeAPNNotification(req PushNotification) {

}

func PushToFCM(req PushNotification) {
}
