package cmd

import (
	"fmt"
	"net/http"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"gopkg.in/gin-gonic/gin.v1"
	"github.com/facebookgo/grace/gracehttp"
)

var (
	serverPort int
	serverInterface string
	useTLS bool
	fullChain string
	privKey string
)

var probesTable = 'registered_probes'

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

func initDatabase() (db *sql.DB, err error) {
	db, err := sql.Open("postgres", postgresDBURI)
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

	/*
	Location struct {
		Latitude `json:"latitude"`,
		Longitude `json:"longitude"`
	} `json:"location"`
	*/

	ProbeFamily string `json:"probe_family"`
	ProbeID string `json:"probe_id"`
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
	
	DeviceToken string `json:"device_token"`

	/*
	Location struct {
		Latitude `json:"latitude"`,
		Longitude `json:"longitude"`
	} `json:"location"`
	*/

	ProbeFamily string `json:"probe_family"`
	ProbeID string `json:"probe_id"`
}

type HeartbeatReq struct {
	MetadataID string `json:"metadata_id"`
	LastHeartbeat string `json:""`
}

type InvalidRequest struct {
	ErrorMessage string
	ErrorCode int
	StatusCode int
}

func Register(db *sql.DB, req RegisterReq) (metadataID string, err error) {
	if (registerReq.Platform == 'ios' && !registerReq.DeviceToken) {
		return nil, InvalidRequest{ErrorMessage: "missing device token",
									ErrorCode: 1,
									StatusCode: http.StatusBadRequest}
	}
	const query = fmt.Sprintf(`INSERT INTO %s (metadata_id,
	probe_cc, probe_asn,
	platform, software_name,
	software_version, supported_tests, network_type
	available_bandwidth, device_token, probe_family,
	probe_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
	probesTable)

	_, err := db.Exec(query, metadataID, req.ProbeCC, req.ProbeASN,
						req.SoftwareName,
						req.SoftwareVersion, pq.Array(req.SupportedTests),
						req.NetworkType, req.AvailableBandwidth,
						req.DeviceToken, req.ProbeFamily, req.ProbeID)
	if (err != nil) {
		ctx.WithError(err).Error("failed to add data to DB")
		return nil, InvalidRequest{ErrorMessage: "error in adding data",
									ErrorCode: 3,
									StatusCode: http.StatusBadRequest}
	}
}

func start() {
	db, err := initDatabase()
	if (err != nil) {
		return
	}

	router := gin.Default()
	router.POST('/api/v1/register', func(c *gin.Context) {
		var registerReq RegisterReq
		if (c.BindJSON(&registerReq) == nil) {
			metadataID, err := Register(db, registerReq)
			if (err != nil) {
				c.JSON(err.StatusCode,
						gin.H{"error": err.ErrorMessage,
								"error_code": err.ErrorCode})
			} else {
				c.JSON(http.StatusOK, gin.H{"metadata_id": metadataID})
			}
		} else {
			c.JSON(http.StatusBadRequest,
					gin.H{"error": "invalid request",
							"error_code": 2})
		}
	})
	router.POST('/api/v1/update', func(c *gin.Context) {
		var updateReq UpdateReq
	})

	router.POST('/api/v1/heartbeat', func(c *gin.Context) {
		var metadata Metadata
	})

	s := &http.Server{
		Addr: ":8080",
		Handler: router,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	gracehttp.Serve(s)
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
