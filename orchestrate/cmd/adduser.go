package cmd

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	jwt "github.com/hellais/jwt-go"
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
	keypath  string
)

// UserData stores information about a user
type UserData struct {
	Username string
	Password string
	KeyPath  string
}

func loadKey(path string) ([]byte, *rsa.PublicKey, error) {
	keyPEM, err := ioutil.ReadFile(path)
	if err != nil {
		return keyPEM, nil, err
	}
	pubkey, err := jwt.ParseRSAPublicKeyFromPEM(keyPEM)
	return keyPEM, pubkey, err
}

// MakeFingerprint should match the output of:
// openssl pkey -pubin -outform DER -in pubkey | openssl dgst -sha256 -binary | hexdump -ve '/1 "%02x"'
func MakeFingerprint(pubkey *rsa.PublicKey) (string, error) {
	keyDER, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(keyDER)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// AddUser adds a user to the database
func AddUser(db *sqlx.DB, user UserData) error {
	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}

	keyPEM, pubkey, err := loadKey(user.KeyPath)
	if err != nil {
		ctx.WithError(err).Errorf("failed to load the key at %s", user.KeyPath)
		return err
	}
	keyFingerprint, err := MakeFingerprint(pubkey)
	if err != nil {
		ctx.WithError(err).Error("failed to load compute the public key")
		return err
	}
	ctx.Infof("Fingerprint: %s", keyFingerprint)

	var userID int
	{
		query := fmt.Sprintf(`INSERT INTO %s (
			username,
			password_hash,
			last_access, role
		) VALUES (
			$1, $2, NOW(), $3
			) RETURNING id`, pq.QuoteIdentifier(common.AccountsTable))

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

		err = stmt.QueryRow(user.Username,
			string(passwordHash),
			"admin").Scan(&userID)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into accounts table, rolling back")
			return err
		}
	}

	{
		query := `INSERT INTO account_keys (
			account_id,
			key_fingerprint,
			key_data
		) VALUES (
			$1, $2, $3
		)`

		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare accounts query")
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(
			userID,
			keyFingerprint,
			keyPEM)
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
		if username == "" {
			ctx.Error("username cannot be empty")
			return
		}
		if password == "" {
			ctx.Error("password cannot be empty")
			return
		}
		if keypath == "" {
			ctx.Error("key cannot be empty")
			return
		}
		user := UserData{
			Username: username,
			Password: password,
			KeyPath:  keypath,
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
	addUserCmd.PersistentFlags().StringVarP(&keypath, "key", "k", "", "specify the path to the key file")
	addUserCmd.MarkFlagRequired("username")
	addUserCmd.MarkFlagRequired("password")
	addUserCmd.MarkFlagRequired("key")
}
