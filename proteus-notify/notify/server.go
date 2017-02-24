package notify

import (
	"fmt"
	"net/http"
	"database/sql"

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

	ctx.Infof("ENV: %s", viper.GetString("environment"))
	if viper.GetString("environment") != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

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
