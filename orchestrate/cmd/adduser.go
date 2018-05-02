package cmd

import (
	"fmt"
	"time"

	common "github.com/ooni/orchestra/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	username string
	password string
	keyid    string
)

// UserData stores information about a user
type UserData struct {
	Username string
	Password string
	KeyID    string
}

// AddUser adds a user to the database
func AddUser(db *sqlx.DB, user UserData) error {
	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}

	{
		query := fmt.Sprintf(`INSERT INTO %s (
			username,
			password_hash,
			keyid,
			last_access,
			role
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5)`,
			pq.QuoteIdentifier(common.AccountsTable))

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.WithError(err).Error("failed to hash password")
			return err
		}

		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare accounts query")
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(user.Username,
			string(passwordHash),
			user.KeyID,
			time.Now().UTC(),
			"admin")
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into accounts table, rolling back")
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return err
	}

	return nil
}

// addUserCmd represents the sign command
var addUserCmd = &cobra.Command{
	Use:   "adduser",
	Short: "Used to add admin users to the orchestra",
	Long:  `This command is used to add users to the database`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := sqlx.Open("postgres", viper.GetString("database.url"))
		if err != nil {
			ctx.WithError(err).Error("failed to connect to the database")
			return
		}
		user := UserData{
			Username: username,
			Password: password,
			KeyID:    keyid,
		}
		err = AddUser(db, user)
		if err != nil {
			ctx.WithError(err).Error("failed to connect to add the user")
			return
		}
		ctx.Infof("Added user \"%s\"", user.Username)
		return
	},
}

func init() {
	RootCmd.AddCommand(addUserCmd)

	// Here you will define your flags and configuration settings.
	addUserCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "specify the username to add")
	addUserCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "specify the username password")
	addUserCmd.PersistentFlags().StringVarP(&keyid, "keyid", "k", "", "specify the keyid of the user")
	addUserCmd.MarkFlagRequired("username")
	addUserCmd.MarkFlagRequired("password")
	addUserCmd.MarkFlagRequired("keyid")
}
