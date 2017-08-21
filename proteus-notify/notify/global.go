package notify

import (
	"crypto/tls"

	"github.com/apex/log"
	apns "github.com/sideshow/apns2"
)

var (
	ctx = log.WithFields(log.Fields{
		"cmd": "notify",
	})
	CertificatePemIos tls.Certificate
	QueueNotification chan PushNotification
	ApnsClient        *apns.Client
)
