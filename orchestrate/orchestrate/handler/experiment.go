package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/gin-gonic/gin"
)

// ExperimentData contains the data for the experiment to be scheduled
type ExperimentData struct {
	Issuer    string                   `form:"iss" json:"iss"`
	ExpiresAt int64                    `form:"exp" json:"exp"`
	ProbeCC   string                   `form:"probe_cc" json:"probe_cc"`
	TestName  string                   `form:"test_name" json:"test_name"`
	Args      []map[string]interface{} `form:"args" json:"args"`
}

// MakeExperiment is used to create a base64 experiment to be signed
func MakeExperiment(c *gin.Context) {
	var exp ExperimentData

	if err := c.Bind(&exp); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	if exp.ExpiresAt == 0 {
		// Default to 1 week expiry time
		exp.ExpiresAt = time.Now().Add(time.Hour * 24 * 7).Unix()
	}

	// XXX we probably want to set the issuer from the auth middleware
	if exp.Issuer == "" {
		exp.Issuer = "testing"
	}

	expData, err := json.Marshal(exp)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	log.Debugf("Serialized: %s", expData)

	b64Str := base64.StdEncoding.EncodeToString(expData)
	c.JSON(http.StatusOK,
		gin.H{"base64": b64Str})
	return
}
