package cmd

import (
	"fmt"

	"github.com/apex/log"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/common"
	"github.com/ooni/orchestra/orchestrate/orchestrate/sched"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initJobDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", viper.GetString("database.url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
}

// startCmd represents the start command
var updateNotify = &cobra.Command{
	Use:   "updatenotify",
	Short: "Send a notification to  update",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := initJobDB()
		if err != nil {
			log.WithError(err).Error("failed to init jdb")
			return
		}
		alertData := sched.AlertData{
			Message: "You may be running an out of date version of OONI Probe which includes a critical bug. Please update to the latest version.",
		}

		query := fmt.Sprintf("SELECT id, token, platform FROM %s",
			pq.QuoteIdentifier(common.ActiveProbesTable))
		query += " WHERE is_token_expired = false AND token != ''"
		query += " AND software_version = '2.0.0'"
		query += " AND software_name = 'ooniprobe-android'"

		rows, err := db.Query(query)
		if err != nil {
			ctx.WithError(err).Error("failed to find targets")
			return
		}
		defer rows.Close()
		for rows.Next() {
			var (
				clientID string
				token    string
				plat     string
			)
			err = rows.Scan(&clientID, &token, &plat)
			if err != nil {
				ctx.WithError(err).Error("failed to iterate over targets")
				return
			}
			target := sched.NewJobTarget(clientID, token, plat, nil, nil, &alertData)

			err = sched.NotifyGorush(
				viper.GetString("core.gorush-url"),
				target)
			if err != nil {
				ctx.WithError(err).Errorf("failed to notify cid: %s", clientID)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(updateNotify)
}
