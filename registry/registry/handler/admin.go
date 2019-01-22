package handler

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	common "github.com/ooni/orchestra/common"
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

func getClientCount(db *sqlx.DB) (int64, error) {
	var count int64

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s",
		pq.QuoteIdentifier(common.ActiveProbesTable))

	row := db.QueryRow(query)
	err := row.Scan(&count)
	if err != nil {
		ctx.WithError(err).Error("failed to list clients")
		return count, err
	}
	return count, err
}

// CountryCount is the count of active probes for the given country
type CountryCount struct {
	CountryCode string `json:"probe_cc"`
	Count       int64  `json:"count"`
}

func getClientCountries(db *sqlx.DB) ([]CountryCount, error) {
	var err error

	countryCounts := make([]CountryCount, 0)

	query := fmt.Sprintf(`SELECT COUNT(*), probe_cc
		FROM %s
		GROUP BY probe_cc`,
		pq.QuoteIdentifier(common.ActiveProbesTable))

	rows, err := db.Query(query)
	if err != nil {
		ctx.WithError(err).Error("failed to list clients")
		return countryCounts, err
	}
	defer rows.Close()
	for rows.Next() {
		var cc CountryCount
		err = rows.Scan(&cc.Count, &cc.CountryCode)
		if err != nil {
			ctx.WithError(err).Error("failed to get client count row")
			continue
		}
		countryCounts = append(countryCounts, cc)
	}
	return countryCounts, err
}

func getResultCount(q ClientsQuery, db *sqlx.DB) (int64, error) {
	var (
		count int64
		args  []interface{}
		err   error
	)

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`,
		pq.QuoteIdentifier(common.ActiveProbesTable))
	query, args = filterClients(q, query, args)

	stmt, err := db.Prepare(query)
	if err != nil {
		ctx.WithError(err).Error("failed prepare result count")
		return count, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(args...)
	err = row.Scan(&count)
	if err != nil {
		ctx.WithError(err).Error("failed to get result count")
		return count, err
	}

	return count, err
}

func filterClients(q ClientsQuery, query string, args []interface{}) (string, []interface{}) {
	if q.CountryCode != "" {
		query += " AND probe_cc = ANY($3)"
		args = append(args, pq.StringArray(
			common.MapToUppercase(strings.Split(q.CountryCode, ","))))
	}
	return query, args
}

// ListClients lists all the clients in the database
func ListClients(db *sqlx.DB, q ClientsQuery) ([]ActiveClient, error) {
	var (
		activeClients []ActiveClient
		args          []interface{}
	)
	args = append(args, q.Limit)
	args = append(args, q.Offset)

	query := fmt.Sprintf(`SELECT
			id, creation_time,
			last_updated,
			probe_cc, probe_asn,
			platform, software_name,
			software_version, supported_tests,
			network_type, available_bandwidth,
			lang_code,
			token, probe_family,
			probe_id
			FROM %s`,
		pq.QuoteIdentifier(common.ActiveProbesTable))

	query, args = filterClients(q, query, args)
	query += " ORDER BY creation_time ASC LIMIT $1 OFFSET $2"

	stmt, err := db.Prepare(query)
	if err != nil {
		ctx.WithError(err).Error("failed prepare clients query")
		return activeClients, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		ctx.WithError(err).Error("failed to list clients")
		return activeClients, err
	}
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

// ClientsQuery this is the query for the clients to be retrieved from the DB
type ClientsQuery struct {
	Limit       int64  `form:"limit" binding:"max=1000"`
	Offset      int64  `form:"offset"`
	CountryCode string `form:"country_code"`
}

// Dict a dictionary datatype
type Dict map[string]interface{}

// MakeMetadata generates the metadata for the request
func (q ClientsQuery) MakeMetadata(db *sqlx.DB) (Dict, error) {
	var err error

	metadata := make(Dict)
	clientCountries, err := getClientCountries(db)
	if err != nil {
		return metadata, err
	}
	metadata["client_countries"] = clientCountries

	clientCount, err := getClientCount(db)
	if err != nil {
		return metadata, err
	}
	metadata["total_client_count"] = clientCount
	resultCount, err := getResultCount(q, db)

	metadata["count"] = resultCount
	metadata["current_page"] = int64(math.Ceil(float64(q.Offset)/float64(q.Limit))) + 1
	metadata["limit"] = q.Limit
	metadata["pages"] = int64(math.Ceil(float64(resultCount) / float64(q.Limit)))
	metadata["next_url"] = fmt.Sprintf("/api/v1/admin/clients?country_code=%s&limit=%d&offset=%d",
		q.CountryCode, q.Limit, q.Offset+q.Limit)
	return metadata, err
}

// ListClientsHandler is the admin handler for listing registered clients
func ListClientsHandler(c *gin.Context) {
	var (
		err          error
		clientsQuery ClientsQuery
	)
	// This is equivalent to setting the default value
	clientsQuery.Limit = 100
	clientsQuery.Offset = 0

	db := c.MustGet("DB").(*sqlx.DB)

	// XXX maybe we can make this stricter by calling c.BindQuery, but that has
	// yet to land in a stable release of gin.
	// See: https://github.com/gin-gonic/gin/pull/1029
	if err = c.Bind(&clientsQuery); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	clientList, err := ListClients(db, clientsQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	metadata, err := clientsQuery.MakeMetadata(db)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"results":  clientList,
			"metadata": metadata,
		})
}
