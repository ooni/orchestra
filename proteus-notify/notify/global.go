package notify

import (
	"github.com/apex/log"
)

var (
	ctx = log.WithFields(log.Fields{
		"cmd": "notify",
	})
	QueueNotification chan PushNotification
	ApnsClient *apns.Client
)
