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
	_ "github.com/facebookgo/grace/gracehttp"
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

type RegisterReq struct {
	ProbeCC string `json:"probe_cc" binding:"required"`
	ProbeASN string `json:"probe_asn" binding:"required"`
	Platform string `json:"platform" binding:"required"`

	SoftwareName string `json:"software_name" binding:"required"`
	SoftwareVersion string `json:"software_version" binding:"required"`
	SupportedTests []string `json:"supported_tests"`

	NetworkType string `json:"network_type"`
	AvailableBandwidth string `json:"available_bandwidth"`
	
	DeviceToken string `json:"device_token"`

	ProbeFamily string `json:"probe_family"`
	ProbeID string `json:"probe_id"`
}

type UpdateReq struct {
	MetadataID string `json:"metadata_id" binding:"required"`
	ProbeCC string `json:"probe_cc"`
	ProbeASN string `json:"probe_asn"`
	Platform string `json:"platform"`

	SoftwareName string `json:"software_name"`
	SoftwareVersion string `json:"software_version"`
	SupportedTests []string `json:"supported_tests"`

	NetworkType string `json:"network_type"`
	AvailableBandwidth string `json:"available_bandwidth"`
	
	DeviceToken string `json:"device_token"`

	ProbeFamily string `json:"probe_family"`
	ProbeID string `json:"probe_id"`
}

type HeartbeatReq struct {
	MetadataID string `json:"metadata_id"`
	LastHeartbeat string `json:""`
}

func Update(db *sql.DB, req UpdateReq) (error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		id, update_time,
		metadata_id,
		probe_cc, probe_asn,
		platform, software_name,
		software_version, supported_tests,
		network_type, available_bandwidth,
		device_token, probe_family,
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

	id := uuid.NewV4().String()
	_, err := db.Exec(query,
						id, time.Now().UTC(),
						req.MetadataID,
						req.ProbeCC, req.ProbeASN,
						req.Platform, req.SoftwareName,
						req.SoftwareVersion, pq.Array(req.SupportedTests),
						req.NetworkType, req.AvailableBandwidth,
						req.DeviceToken, req.ProbeFamily,
						req.ProbeID, "update")
	if (err != nil) {
		ctx.WithError(err).Error("failed to add data to update table")
		return errors.New("error in adding data to update probes")
	}
	return nil
}

func Register(db *sql.DB, req RegisterReq) (string, error) {
	if (req.Platform == "ios" && req.DeviceToken == "") {
		return "", errors.New("missing device token")
	}

	// XXX Wrap this all in a transaction
	// figure out what this means: https://github.com/lib/pq/issues/81
	query := fmt.Sprintf(`INSERT INTO %s (
		id, creation_time,
		last_updated,
		probe_cc, probe_asn,
		platform, software_name,
		software_version, supported_tests,
		network_type, available_bandwidth,
		device_token, probe_family,
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

	metadataID := uuid.NewV4().String()
	_, err := db.Exec(query,
						metadataID, time.Now().UTC(),
						time.Now().UTC(),
						req.ProbeCC, req.ProbeASN,
						req.Platform, req.SoftwareName,
						req.SoftwareVersion, pq.Array(req.SupportedTests),
						req.NetworkType, req.AvailableBandwidth,
						req.DeviceToken, req.ProbeFamily,
						req.ProbeID)
	if (err != nil) {
		ctx.WithError(err).Error("failed to add data to active probe table")
		return "", errors.New("error in adding data to active probes")
	}

	query = fmt.Sprintf(`INSERT INTO %s (
		id, update_time,
		metadata_id,
		probe_cc, probe_asn,
		platform, software_name,
		software_version, supported_tests,
		network_type, available_bandwidth,
		device_token, probe_family,
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

	updateID := uuid.NewV4().String()
	_, err = db.Exec(query,
						updateID, time.Now().UTC(),
						metadataID,
						req.ProbeCC, req.ProbeASN,
						req.Platform, req.SoftwareName,
						req.SoftwareVersion, pq.Array(req.SupportedTests),
						req.NetworkType, req.AvailableBandwidth,
						req.DeviceToken, req.ProbeFamily,
						req.ProbeID, "register")
	if (err != nil) {
		ctx.WithError(err).Error("failed to add data to update table")
		return "", errors.New("error in adding data to update probes")
	}

	return metadataID, nil
}

func start() {
	db, err := initDatabase()

	if (err != nil) {
		ctx.WithError(err).Error("failed to connect to DB")
		return
	}
	defer db.Close()

	router := gin.Default()
	router.POST("/api/v1/register", func(c *gin.Context) {
		var registerReq RegisterReq
		err := c.BindJSON(&registerReq)
		if (err != nil) {
			ctx.WithError(err).Error("invalid request")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request"})
			return
		}

		metadataID, err := Register(db, registerReq)
		if (err != nil) {
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"metadata_id": metadataID})
		return
	})
	router.POST("/api/v1/update", func(c *gin.Context) {
		var updateReq UpdateReq
		err := c.BindJSON(&updateReq)
		if (err != nil) {
			ctx.WithError(err).Error("invalid request")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request"})
			return
		}
		err = Update(db, updateReq)
		if (err != nil) {
			ctx.WithError(err).Error("failed to update")
			c.JSON(http.StatusBadRequest,
					gin.H{"error": err.Error()})
			return
		}
	})

	router.POST("/api/v1/heartbeat", func(c *gin.Context) {
		//var heartbeatReq HeartbeatReq
	})
	
	Addr := fmt.Sprintf("%s:%d", viper.GetString("server-address"),
								viper.GetInt("server-port"))
	ctx.Infof("starting on %s", Addr)
	s := &http.Server{
		Addr: Addr,
		Handler: router,
	}
	//gracehttp.Serve(s)
	s.ListenAndServe()
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
