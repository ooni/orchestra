package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	common "github.com/thetorproject/proteus/proteus-common"
)

// ActiveClient metadata about an active client
type ActiveClient struct {
	ClientID string `json:"client_id"`

	ProbeCC  string `json:"probe_cc"`
	ProbeASN string `json:"probe_asn"`
	Platform string `json:"platform"`

	SoftwareName    string `json:"software_name"`
	SoftwareVersion string `json:"software_version"`
	SupportedTests  string `json:"supported_tests"`

	NetworkType        string `json:"network_type"`
	AvailableBandwidth string `json:"available_bandwidth"`
	Language           string `json:"language"`

	Token string `json:"token"`

	ProbeFamily string `json:"probe_family"`
	ProbeID     string `json:"probe_id"`

	LastUpdated  time.Time `json:"last_updated"`
	CreationTime time.Time `json:"creation_time"`
}

// ListClients lists all the clients in the database
func ListClients(db *sqlx.DB) ([]ActiveClient, error) {
	var activeClients []ActiveClient
	query := fmt.Sprintf(`SELECT
			id, creation_time,
			last_updated,
			probe_cc, probe_asn,
			platform, software_name,
			software_version, supported_tests,
			network_type, available_bandwidth,
			lang_code,
			token, probe_family,
			probe_id FROM %s`,
		pq.QuoteIdentifier(common.ActiveProbesTable))

	rows, err := db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list clients")
		return activeClients, err
	}
	defer rows.Close()
	for rows.Next() {
		var ac ActiveClient
		/* In theory Language should be a string and we accept it as such
		   when processing incoming JSON. Yet, because in #6 we migrated the
		   schema adding the language column, there are plenty of rows in
		   which the language is actually `null`. This would cause the Scan
		   to fail. Fix passing in a nullable type for language and then
		   setting the proper String type as JSON expects it only _if_
		   the value in the database is not `null`. (I know this creates
		   glue, however I don't want to change the datatype.) */
		var maybeLanguage sql.NullString
		err := rows.Scan(&ac.ClientID,
			&ac.CreationTime,
			&ac.LastUpdated,
			&ac.ProbeCC,
			&ac.ProbeASN,
			&ac.Platform,
			&ac.SoftwareName,
			&ac.SoftwareVersion,
			&ac.SupportedTests,
			&ac.NetworkType,
			&ac.AvailableBandwidth,
			&maybeLanguage,
			&ac.Token,
			&ac.ProbeFamily,
			&ac.ProbeID)
		if err != nil {
			ctx.WithError(err).Error("failed to iterate over clients")
			return activeClients, err
		}
		if maybeLanguage.Valid {
			ac.Language = maybeLanguage.String
		}
		activeClients = append(activeClients, ac)
	}
	return activeClients, nil
}

// ListClientsHandler is the admin handler for listing registered clients
func ListClientsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	clientList, err := ListClients(db)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"active_clients": clientList})
}
