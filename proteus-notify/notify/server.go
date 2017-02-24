package notify

import (
	"fmt"
	"time"
	"errors"
	"net/http"
	"database/sql"

	"github.com/satori/go.uuid"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	"gopkg.in/gin-gonic/gin.v1"
	"github.com/facebookgo/grace/gracehttp"
)

type OEvent struct {
	Name string `json:"name"`
	Filter interface{} `json:"filter"`
	Schedule string `json:"schedule"`
	Delay string `json:"delay"`
	Task interface{} `json:"task"`
}

type NotifyReq struct {
	ClientIDs []string `json:"client_ids" binding:"required"`
	Priority string `json:"priority"`
	Event OEvent `json:"event"`
}

func isClientRegistered(db *sql.DB, clientID string) (bool, error) {
	var found string
	query := fmt.Sprintf(`SELECT id FROM %s WHERE id = $1`,
				pq.QuoteIdentifier(viper.GetString("active-probes-table")))
	err := db.QueryRow(query, clientID).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", viper.GetString("database-url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
}

func StartServer() {
	db, err := initDatabase()

	if (err != nil) {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	router := gin.Default()
	router.POST("/api/v1/notify", func(c *gin.Context) {
		var notifyReq NotifyReq
		err := c.BindJSON(&notifyReq)
		if (err != nil) {
			ctx.WithError(err).Error("invalid request")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request"})
			return
		}

		clientID , err := Register(db, registerReq)
		if (err != nil) {
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"client_id": clientID})
		return
	})

	Addr := fmt.Sprintf("%s:%d", viper.GetString("server-address"),
								viper.GetInt("server-port"))
	ctx.Infof("starting on %s", Addr)
	s := &http.Server{
		Addr: Addr,
		Handler: router,
	}
	gracehttp.Serve(s)
}
