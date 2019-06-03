package cmd

import (
	"bufio"
	"fmt"
	"os"

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
var notify = &cobra.Command{
	Use:   "notify",
	Short: "Send a notification message to some users",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		where := viper.GetString("notify.where")
		message := viper.GetString("notify.message")
		if message == "" {
			log.Error("Message is empty")
			return
		}

		db, err := initJobDB()
		if err != nil {
			log.WithError(err).Error("failed to init jdb")
			return
		}
		// "You may be running an out of date version of OONI Probe which includes a critical bug. Please update to the latest version."
		alertData := sched.AlertData{
			Message: message,
		}

		query := fmt.Sprintf("SELECT id, token, platform FROM %s",
			pq.QuoteIdentifier(common.ActiveProbesTable))
		query += " WHERE is_token_expired = false AND token != '' AND "
		query += where

		rows, err := db.Query(query)
		if err != nil {
			ctx.WithError(err).Error("failed to find targets")
			return
		}
		var targets []*sched.JobTarget
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
			targets = append(targets, sched.NewJobTarget(clientID, token, plat, nil, nil, &alertData))
		}
		reader := bufio.NewReader(os.Stdin)
		log.Infof("You are about to send to %d users the message: \"%s\". Press enter continue or ctrl-c to cancel.", len(targets), message)
		reader.ReadString('\n')

		for _, target := range targets {
			err = sched.NotifyGorush(
				viper.GetString("core.gorush-url"),
				target)
			if err != nil {
				ctx.WithError(err).Errorf("failed to notify cid: %s", target.ClientID)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(notify)

	notify.PersistentFlags().StringP("where", "", "", "Where query statement to select users by, example software_version = '2.0.0'")
	notify.PersistentFlags().StringP("message", "", "", "The content of the message to send")
	viper.BindPFlag("notify.where", notify.PersistentFlags().Lookup("where"))
	viper.BindPFlag("notify.message", notify.PersistentFlags().Lookup("message"))
}
