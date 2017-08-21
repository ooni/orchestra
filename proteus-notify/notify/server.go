package notify

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	"gopkg.in/gin-gonic/gin.v1"
)

type OEvent struct {
	Name     string      `json:"name"`
	Filter   interface{} `json:"filter"`
	Schedule string      `json:"schedule"`
	Delay    string      `json:"delay"`
	Task     interface{} `json:"task"`
}

type NotifyReq struct {
	ClientIDs []string               `json:"client_ids" binding:"required"`
	Priority  string                 `json:"priority"`
	Event     map[string]interface{} `json:"event"`
}

type TokenPlatform struct {
	Token    string
	Platform string
}

func GetClientTokenPlatform(db *sql.DB, clientID string) (TokenPlatform, error) {
	var tp TokenPlatform
	query := fmt.Sprintf(`SELECT token, platform FROM %s WHERE id = $1`,
		pq.QuoteIdentifier(viper.GetString("database.active-probes-table")))
	err := db.QueryRow(query, clientID).Scan(&tp.Token, &tp.Platform)
	ctx.Debugf("found %s, %s", tp.Token, tp.Platform)
	// The caller is responsible for checking the error
	return tp, err
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
}

func MakeNotifications(db *sql.DB, req NotifyReq) ([]PushNotification, error) {
	var notifications []PushNotification

	androidNotification := PushNotification{
		Topic:    viper.GetString("fcm.topic"),
		Priority: req.Priority,
		Retry:    5,
		Platform: "android",
	}
	androidNotification.Data = req.Event
	iosNotification := PushNotification{
		Topic:    viper.GetString("apn.topic"),
		Priority: req.Priority,
		Retry:    5,
		Platform: "ios",
	}
	iosNotification.Data = req.Event

	for _, clientID := range req.ClientIDs {
		tp, err := GetClientTokenPlatform(db, clientID)
		if err == sql.ErrNoRows {
			ctx.Warnf("could not find client with ID %s. ignoring", clientID)
			continue
		}

		if err != nil {
			ctx.WithError(err).Error("failed to lookup token")
			continue
		}

		if tp.Platform == "ios" {
			ctx.Debugf("appending token %s", tp.Token)
			iosNotification.Tokens = append(iosNotification.Tokens, tp.Token)
		} else if tp.Platform == "android" {
			androidNotification.Tokens = append(androidNotification.Tokens,
				tp.Token)
		} else {
			ctx.Warnf("unsupported platform type %s", tp.Platform)
		}
	}
	if len(androidNotification.Tokens) > 0 {
		notifications = append(notifications, androidNotification)
	}
	if len(iosNotification.Tokens) > 0 {
		notifications = append(notifications, iosNotification)
	}
	return notifications, nil
}

func StartServer() {
	err := InitApnsClient()
	if err != nil {
		ctx.WithError(err).Error("failed to connect to APN")
		return
	}

	InitWorkers(viper.GetInt("core.worker-num"),
		viper.GetInt("core.queue-size"))

	db, err := initDatabase()

	if err != nil {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	ctx.Infof("ENV: %s", viper.GetString("core.environment"))
	if viper.GetString("environment") != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.POST("/api/v1/notify", func(c *gin.Context) {
		var notifyReq NotifyReq
		err := c.BindJSON(&notifyReq)
		if err != nil {
			ctx.WithError(err).Error("invalid request")
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "invalid request"})
			return
		}
		notifications, err := MakeNotifications(db, notifyReq)

		for _, notification := range notifications {
			go PushToAny(notification)
		}

		c.JSON(http.StatusOK,
			gin.H{"status": "ok"})
		return
	})

	Addr := fmt.Sprintf("%s:%d", viper.GetString("api.address"),
		viper.GetInt("api.port"))
	ctx.Infof("starting on %s", Addr)
	s := &http.Server{
		Addr:    Addr,
		Handler: router,
	}
	gracehttp.Serve(s)
}
