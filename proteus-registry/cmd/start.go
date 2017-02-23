package cmd

import (
	"fmt"
	"time"
	"errors"
	"net/http"
	"database/sql"

	"github.com/satori/go.uuid"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/gin-gonic/gin.v1"
	"github.com/facebookgo/grace/gracehttp"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", viper.GetString("database-url"))
	if err != nil {
		ctx.Error("failed to open database")
		return nil, err
	}
	return db, err
}

type UpdateReq struct {
	ProbeCC string `json:"probe_cc"`
	ProbeASN string `json:"probe_asn"`
	Platform string `json:"platform"`

	SoftwareName string `json:"software_name"`
	SoftwareVersion string `json:"software_version"`
	SupportedTests []string `json:"supported_tests"`

	NetworkType string `json:"network_type"`
	AvailableBandwidth string `json:"available_bandwidth"`
	
	Token string `json:"token"`

	ProbeFamily string `json:"probe_family"`
	ProbeID string `json:"probe_id"`
}

func IsClientRegistered(db *sql.DB, clientID string) (bool, error) {
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

func Update(db *sql.DB, clientID string, req UpdateReq) (error) {
	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return err
	}

	// Write into the updates table
	{
		query := fmt.Sprintf(`INSERT INTO %s (
			id, update_time,
			client_id,
			probe_cc, probe_asn,
			platform, software_name,
			software_version, supported_tests,
			network_type, available_bandwidth,
			token, probe_family,
			probe_id, update_type
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6,
			$7, $8,
			$9, $10,
			$11, $12,
			$13, $14, $15)`,
			pq.QuoteIdentifier(viper.GetString("probe-updates-table")))

		stmt, err := tx.Prepare(query)
		if (err != nil) {
			ctx.WithError(err).Error("failed to prepare update probes query")
			return err
		}

		updateID := uuid.NewV4().String()
		_, err = stmt.Exec(updateID, time.Now().UTC(),
							clientID,
							req.ProbeCC, req.ProbeASN,
							req.Platform, req.SoftwareName,
							req.SoftwareVersion, pq.Array(req.SupportedTests),
							req.NetworkType, req.AvailableBandwidth,
							req.Token, req.ProbeFamily,
							req.ProbeID, "register")
		if (err != nil) {
			ctx.WithError(err).Error("failed to add data to update table, rolling back")
			tx.Rollback()
			return errors.New("error in adding data to update probes")
		}
	}

	// Write into the active probes table
	{
		query := fmt.Sprintf(`UPDATE %s SET
			last_updated = $2,
			probe_cc = $3,
			probe_asn = $4,
			platform = $5,
			software_name = $6,
			software_version = $7,
			supported_tests = $8,
			network_type = $9,
			available_bandwidth = $10,
			token = $11,
			probe_family = $12,
			probe_id = $13
			WHERE id = $1`,
			pq.QuoteIdentifier(viper.GetString("active-probes-table")))

		stmt, err := tx.Prepare(query)
		if (err != nil) {
			ctx.WithError(err).Error("failed to prepare update probes query")
			return err
		}
		_, err = stmt.Exec(clientID,
							time.Now().UTC(),
							req.ProbeCC,
							req.ProbeASN,
							req.Platform,
							req.SoftwareName,
							req.SoftwareVersion,
							pq.Array(req.SupportedTests),
							req.NetworkType,
							req.AvailableBandwidth,
							req.Token,
							req.ProbeFamily,
							req.ProbeID)
		if (err != nil) {
			ctx.WithError(err).Error("failed to update active table, rolling back")
			tx.Rollback()
			return errors.New("failed to update active table")
		}
	}

	if err := tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return err
	}

	return nil
}

type RegisterReq struct {
	ProbeCC string `json:"probe_cc" binding:"required"`
	ProbeASN string `json:"probe_asn" binding:"required"`
	Platform string `json:"platform" binding:"required"`

	SoftwareName string `json:"software_name" binding:"required"`
	SoftwareVersion string `json:"software_version" binding:"required"`
	SupportedTests []string `json:"supported_tests"`

	NetworkType string `json:"network_type"`
	AvailableBandwidth string `json:"available_bandwidth"`
	
	Token string `json:"token"`

	ProbeFamily string `json:"probe_family"`
	ProbeID string `json:"probe_id"`
}

func Register(db *sql.DB, req RegisterReq) (string, error) {
	if ((req.Platform == "ios" || req.Platform == "android") && req.Token == "") {
		return "", errors.New("missing device token")
	}

	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open transaction")
		return "", err
	}

	var clientID = uuid.NewV4().String()

	{
		query := fmt.Sprintf(`INSERT INTO %s (
			id, creation_time,
			last_updated,
			probe_cc, probe_asn,
			platform, software_name,
			software_version, supported_tests,
			network_type, available_bandwidth,
			token, probe_family,
			probe_id
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6,
			$7, $8,
			$9, $10,
			$11, $12,
			$13, $14)`,
			pq.QuoteIdentifier(viper.GetString("active-probes-table")))

		stmt, err := tx.Prepare(query)
		if (err != nil) {
			ctx.WithError(err).Error("failed to prepare active probes query")
			return "", err
		}
		defer stmt.Close()

		_, err = stmt.Exec(clientID, time.Now().UTC(),
							time.Now().UTC(),
							req.ProbeCC, req.ProbeASN,
							req.Platform, req.SoftwareName,
							req.SoftwareVersion, pq.Array(req.SupportedTests),
							req.NetworkType, req.AvailableBandwidth,
							req.Token, req.ProbeFamily,
							req.ProbeID)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into active probes table, rolling back")
			return "", err
		}
	}

	{
		query := fmt.Sprintf(`INSERT INTO %s (
			id, update_time,
			client_id,
			probe_cc, probe_asn,
			platform, software_name,
			software_version, supported_tests,
			network_type, available_bandwidth,
			token, probe_family,
			probe_id, update_type
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6,
			$7, $8,
			$9, $10,
			$11, $12,
			$13, $14, $15)`,
			pq.QuoteIdentifier(viper.GetString("probe-updates-table")))

		stmt, err := tx.Prepare(query)
		if (err != nil) {
			ctx.WithError(err).Error("failed to prepare update probes query")
			return "", err
		}
		defer stmt.Close()

		updateID := uuid.NewV4().String()
		_, err = stmt.Exec(updateID, time.Now().UTC(),
							clientID,
							req.ProbeCC, req.ProbeASN,
							req.Platform, req.SoftwareName,
							req.SoftwareVersion, pq.Array(req.SupportedTests),
							req.NetworkType, req.AvailableBandwidth,
							req.Token, req.ProbeFamily,
							req.ProbeID, "register")
		if (err != nil) {
			ctx.WithError(err).Error("failed to add data to update table, rolling back")
			tx.Rollback()
			return "", errors.New("error in adding data to update probes")
		}
	}

	if err := tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return "", err
	}

	return clientID, nil
}

func start() {
	db, err := initDatabase()

	if (err != nil) {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	router := gin.Default()
	router.POST("/api/v1/clients", func(c *gin.Context) {
		var registerReq RegisterReq
		err := c.BindJSON(&registerReq)
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

	// XXX do we also want to support a PATCH method?
	router.PUT("/api/v1/clients/:client_id", func(c *gin.Context) {
		var updateReq UpdateReq
		clientID := c.Param("client_id")
		err := c.BindJSON(&updateReq)
		if (err != nil) {
			ctx.WithError(err).Error("invalid request")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request"})
			return
		}
		isRegistered, err := IsClientRegistered(db, clientID)
		if (err != nil) {
			ctx.WithError(err).Error("failed to learn if client is registered")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}
		if (isRegistered == false) {
			c.JSON(http.StatusNotFound,
					gin.H{"error": "client is not registered"})
			return
		}

		err = Update(db, clientID, updateReq)
		if (err != nil) {
			ctx.WithError(err).Error("failed to update")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}
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

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	startCmd.PersistentFlags().IntP("server-port", "", 8080, "Which port we should bind to")
	startCmd.PersistentFlags().StringP("server-address", "", "127.0.0.1", "Which interface we should listen on")
	viper.BindPFlag("server-port", startCmd.PersistentFlags().Lookup("server-port"))
	viper.BindPFlag("server-address", startCmd.PersistentFlags().Lookup("server-address"))
}
