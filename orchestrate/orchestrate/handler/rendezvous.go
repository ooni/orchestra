package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	common "github.com/ooni/orchestra/common"
)

// DomainFrontedCollector is a {"domain": "a", "front": "b"} map
type DomainFrontedCollector struct {
	Domain string `json:"domain"`
	Front  string `json:"front"`
}

// CollectorInfo holds the type and address of a collector
type CollectorInfo struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

// GetCollectors returns the list of collectors available
func GetCollectors(types string, db *sqlx.DB) ([]CollectorInfo, error) {
	var (
		err  error
		args []interface{}
	)
	collectors := make([]CollectorInfo, 0)

	query := fmt.Sprintf(`SELECT
		type,
		address,
		front_domain
		FROM %s`,
		pq.QuoteIdentifier(common.CollectorsTable))
	if types != "" {
		query += " WHERE type = ANY($1)"
		args = append(args, pq.StringArray(strings.Split(types, ",")))
	}
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return collectors, nil
		}
		ctx.WithError(err).Error("failed to get collectors")
		return collectors, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			ctype    string
			caddress string
			cfront   sql.NullString
		)
		err = rows.Scan(&ctype, &caddress, &cfront)
		if err != nil {
			ctx.WithError(err).Error("failed to get collector row")
			// In this case we fail fast
			return collectors, err
		}
		if ctype == "domain_fronted" {
			if !cfront.Valid {
				ctx.Error("domain_fronted collector with bad front domain")
				return collectors, err
			}
			caddress = fmt.Sprintf("%s@%s", caddress, cfront.String)
		}
		collectors = append(collectors, CollectorInfo{
			Type:    ctype,
			Address: caddress,
		})
	}
	return collectors, nil
}

// TestHelperInfo holds the name, type and address of a test helper
type TestHelperInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
}

// GetTestHelpers returns a list of test helpers
func GetTestHelpers(names string, db *sqlx.DB) ([]TestHelperInfo, error) {
	var (
		err  error
		args []interface{}
	)
	testHelpers := make([]TestHelperInfo, 0)
	query := fmt.Sprintf(`SELECT
		name,
		type,
		address
		FROM %s`,
		pq.QuoteIdentifier(common.TestHelpersTable))
	if names != "" {
		query += " WHERE name = ANY($1)"
		args = append(args, pq.StringArray(strings.Split(names, ",")))
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return testHelpers, nil
		}
		ctx.WithError(err).Error("failed to get test helpers")
		return testHelpers, err
	}
	defer rows.Close()
	for rows.Next() {
		var testHelper TestHelperInfo
		err = rows.Scan(&testHelper.Name, &testHelper.Type, &testHelper.Address)
		if err != nil {
			ctx.WithError(err).Error("failed to get test_helper row")
			continue
		}
		testHelpers = append(testHelpers, testHelper)
	}
	return testHelpers, nil
}

// CollectorsHandler returns the list of requested collectors
func CollectorsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	types := c.Query("types")
	collectors, err := GetCollectors(types, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"results": collectors})
	return
}

// TestHelpersHandler returns the list of requested test helpers
func TestHelpersHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	names := c.Query("names")
	testHelpers, err := GetTestHelpers(names, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"results": testHelpers})
	return
}
