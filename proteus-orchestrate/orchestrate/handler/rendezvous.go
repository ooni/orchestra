package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	common "github.com/thetorproject/proteus/proteus-common"
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

// GetCollectors returns a map of collectors keyed by their type
func GetCollectors(db *sqlx.DB) ([]CollectorInfo, error) {
	var (
		err error
	)
	collectors := make([]CollectorInfo, 0)
	query := fmt.Sprintf(`SELECT
		type,
		address,
		front_domain
		FROM %s`,
		pq.QuoteIdentifier(common.CollectorsTable))
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
			// XXX should we fail fast? I think so
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

// GetTestHelpers returns a map of test helpers keyed by the test name
func GetTestHelpers(db *sqlx.DB) ([]TestHelperInfo, error) {
	var (
		err error
	)
	testHelpers := make([]TestHelperInfo, 0)
	query := fmt.Sprintf(`SELECT
		test_name,
		address
		FROM %s`,
		pq.QuoteIdentifier(common.TestHelpersTable))
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return testHelpers, nil
		}
		ctx.WithError(err).Error("failed to get test helpers")
		return testHelpers, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			testName string
			address  string
		)
		err = rows.Scan(&testName, &address)
		if err != nil {
			ctx.WithError(err).Error("failed to get test_helper row")
			continue
		}
		testHelpers = append(testHelpers, TestHelperInfo{
			Name:    testName,
			Address: address,
		})
	}
	return testHelpers, nil
}

// upperAndWhitelist checks if a list of strings are uppercased and inside the
// list, returns the list with only the items present in the whitelist
func mapToUppercase(vs []string) []string {
	vso := make([]string, len(vs))
	for i, v := range vs {
		vso[i] = strings.ToUpper(v)
	}
	return vso
}

// prepareURLsQuery returns the statement to get all the inputs for the
// given countries and category codes
func prepareURLsQuery(q URLsQuery, db *sqlx.DB) (*sql.Stmt, []interface{}, error) {
	var (
		countryCodes []string
		args         []interface{}
	)
	args = append(args, q.Limit)
	countryCodes = append(countryCodes, "XX")
	if q.CountryCode != "" {
		countryCodes = append(countryCodes, strings.ToUpper(q.CountryCode))
	}

	query := fmt.Sprintf(`SELECT
		url,
		cat_code,
		alpha_2
		FROM %s urls
		INNER JOIN %s countries ON urls.country_no = countries.country_no
		INNER JOIN %s url_cats ON urls.cat_no = url_cats.cat_no`,
		pq.QuoteIdentifier(common.URLsTable),
		pq.QuoteIdentifier(common.CountriesTable),
		pq.QuoteIdentifier(common.URLCategoriesTable))
	// countries is always greater than zero
	query += " WHERE alpha_2 = ANY($2)"
	args = append(args, pq.StringArray(countryCodes))
	if q.CategoryCodes != "" {
		query += " AND cat_code = ANY($3)"
		args = append(args, pq.StringArray(
			mapToUppercase(strings.Split(q.CategoryCodes, ","))))
	}
	query += " ORDER BY random() LIMIT $1"
	stmt, err := db.Prepare(query)
	return stmt, args, err
}

// URLInfo holds the name, type and address of a test helper
type URLInfo struct {
	CategoryCode string `json:"category_code"`
	URL          string `json:"url"`
	CountryCode  string `json:"country_code"`
}

// GetURLs returns a slice of test inputs
func GetURLs(q URLsQuery, db *sqlx.DB) ([]URLInfo, error) {
	var (
		err error
	)
	urls := make([]URLInfo, 0)
	stmt, args, err := prepareURLsQuery(q, db)
	if err != nil {
		ctx.WithError(err).Error("failed to prepare query")
		return urls, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Debugf("got an empty result")
			return urls, nil
		}
		ctx.WithError(err).Error("failed to get test inputs (urls)")
		return urls, err
	}
	for rows.Next() {
		var url URLInfo
		err = rows.Scan(&url.URL, &url.CategoryCode, &url.CountryCode)
		if err != nil {
			ctx.WithError(err).Error("failed to get test input row (urls)")
			continue
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// URLsQuery is the user issued request for URLs
type URLsQuery struct {
	Limit         int64  `form:"limit" binding:"max=1000"`
	CountryCode   string `form:"country_code"`
	CategoryCodes string `form:"category_codes"`
}

// MakeMetadata generates the metadata for the request
func (q URLsQuery) MakeMetadata() map[string]interface{} {
	// XXX populate this with real data
	return map[string]interface{}{
		"count":        -1,
		"current_page": -1,
		"limit":        q.Limit,
		"pages":        -1,
		"next_url":     "",
	}
}

func validateCSVMapStr(csvStr string, m mapStrStruct) bool {
	if csvStr == "" {
		// It's ok if the value is missing
		return true
	}
	for _, v := range strings.Split(csvStr, ",") {
		_, present := m[strings.ToUpper(v)]
		if !present {
			return false
		}
	}
	return true
}

// URLsHandler returns the list of requested URLs
func URLsHandler(c *gin.Context) {
	var (
		err       error
		urlsQuery URLsQuery
	)
	// This is equivalent to setting the default value
	urlsQuery.Limit = 100

	if validateCSVMapStr(urlsQuery.CountryCode, allCountryCodes) == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country_code"})
		return
	}
	if validateCSVMapStr(urlsQuery.CategoryCodes, allCategoryCodes) == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_codes"})
		return
	}

	db := c.MustGet("DB").(*sqlx.DB)

	// XXX maybe we can make this stricter by calling c.BindQuery, but that has
	// yet to land in a stable release of gin.
	// See: https://github.com/gin-gonic/gin/pull/1029
	if err = c.Bind(&urlsQuery); err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	urls, err := GetURLs(urlsQuery, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}

	metadata := urlsQuery.MakeMetadata()
	c.JSON(http.StatusOK,
		gin.H{
			"metadata": metadata,
			"results":  urls,
		})
	return
}

// CollectorsHandler returns the list of requested collectors
func CollectorsHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	collectors, err := GetCollectors(db)
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

	testHelpers, err := GetTestHelpers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"results": testHelpers})
	return
}
