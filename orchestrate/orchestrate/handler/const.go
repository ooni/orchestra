package handler

import "github.com/apex/log"

var ctx = log.WithFields(log.Fields{
	"pkg": "handler",
	"cmd": "ooni-orchestrate",
})
