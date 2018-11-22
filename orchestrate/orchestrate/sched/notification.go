package sched

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

// NotifyGorush tell gorush to notify clients
func NotifyGorush(notification *GoRushNotification) error {
	var (
		err error
	)

	path, _ := url.Parse("/api/push")
	baseURL, err := url.Parse(viper.GetString("core.gorush-url"))
	if err != nil {
		return err
	}

	notifyReq := GoRushReq{
		Notifications: []*GoRushNotification{notification},
	}

	jsonStr, err := json.Marshal(notifyReq)
	if err != nil {
		ctx.WithError(err).Error("failed to marshal data")
		return err
	}
	u := baseURL.ResolveReference(path)
	ctx.Debugf("sending notify request: %s", jsonStr)
	req, err := http.NewRequest("POST",
		u.String(),
		bytes.NewBuffer(jsonStr))
	if err != nil {
		ctx.WithError(err).Error("failed to send request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if viper.IsSet("auth.gorush-basic-auth-user") {
		req.SetBasicAuth(viper.GetString("auth.gorush-basic-auth-user"),
			viper.GetString("auth.gorush-basic-auth-password"))
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ctx.WithError(err).Error("http request failed")
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.WithError(err).Error("failed to read response body")
		return err
	}
	ctx.Debugf("got response: %s", body)
	// XXX do we also want to check the body?
	if resp.StatusCode != 200 {
		ctx.Debugf("got invalid status code: %d", resp.StatusCode)
		return errors.New("http request returned invalid status code")
	}
	return nil
}

// NotifyReq is the reuqest for sending this particular notification message
// XXX this is duplicated in proteus-notify
type NotifyReq struct {
	ClientIDs []string               `json:"client_ids"`
	Event     map[string]interface{} `json:"event"`
}

// GoRushNotification all the notification metadata for gorush
type GoRushNotification struct {
	Tokens           []string               `json:"tokens"`
	Platform         int                    `json:"platform"`
	Message          string                 `json:"message"`
	Topic            string                 `json:"topic"`
	To               string                 `json:"to"`
	Data             map[string]interface{} `json:"data"`
	ContentAvailable bool                   `json:"content_available"`
	Notification     map[string]string      `json:"notification"`
}

// GoRushReq a wrapper for a gorush notification request
type GoRushReq struct {
	Notifications []*GoRushNotification `json:"notifications"`
}

// MakeAlertNotification will make a Job and JobTarget into a GoRushNotification
func MakeAlertNotification(j *Job, jt *JobTarget) (*GoRushNotification, error) {
	var notificationType = "default"
	notification := &GoRushNotification{
		Tokens: []string{jt.Token},
	}
	ctx.Debugf("making alert data for %v", j)
	alertData := j.Data.(*AlertData)
	if _, ok := alertData.Extra["href"]; ok {
		notificationType = "open_href"
	}
	notification.Message = alertData.Message
	notification.Data = map[string]interface{}{
		"type":    notificationType,
		"payload": alertData.Extra,
	}
	notification.Notification = make(map[string]string)

	if jt.Platform == "ios" {
		notification.Platform = 1
		notification.Topic = viper.GetString("core.notify-topic-ios")
	} else if jt.Platform == "android" {
		notification.Notification["click_action"] = viper.GetString(
			"core.notify-click-action-android")
		notification.Platform = 2
		/* We don't need to send a topic on Android. As the response message of
		   failed requests say: `Must use either "registration_ids" field or
		   "to", not both`. And we need `registration_ids` because we send in
		   multicast to many clients. More evidence, as usual, on SO:
		   <https://stackoverflow.com/a/33440105>. */
	} else {
		return nil, ErrUnsupportedPlatform
	}
	return notification, nil
}

// ErrUnsupportedPlatform when the platform is not supported
var ErrUnsupportedPlatform = errors.New("unsupported platform")

// MakeExperimentNotification makes a GorushNotification for a given JobTarget and Experiment
func MakeExperimentNotification(j *Job, jt *JobTarget, expID string) (*GoRushNotification, error) {
	notification := &GoRushNotification{
		Tokens: []string{jt.Token},
	}
	notification.Data = map[string]interface{}{
		"type": "run_experiment",
		"payload": map[string]string{
			"experiment_id": expID,
		},
	}
	notification.ContentAvailable = true
	notification.Notification = make(map[string]string)

	if jt.Platform == "ios" {
		notification.Platform = 1
		notification.Topic = viper.GetString("core.notify-topic-ios")
	} else if jt.Platform == "android" {
		notification.Platform = 2
		/* We don't need to send a topic on Android. As the response message of
		   failed requests say: `Must use either "registration_ids" field or
		   "to", not both`. And we need `registration_ids` because we send in
		   multicast to many clients. More evidence, as usual, on SO:
		   <https://stackoverflow.com/a/33440105>. */
	} else {
		return nil, ErrUnsupportedPlatform
	}
	return notification, nil
}
