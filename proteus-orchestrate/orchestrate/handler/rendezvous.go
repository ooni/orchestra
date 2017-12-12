package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	common "github.com/thetorproject/proteus/proteus-common"
)

// upperAndWhitelist checks if a list of strings are uppercased and inside the
// list, returns the list with only the items present in the whitelist
func upperAndWhitelist(ins []string, whitelist mapStrStruct) ([]string, error) {
	outs := make([]string, len(ins))
	for i, v := range ins {
		outs[i] = strings.ToUpper(v)
		_, present := whitelist[outs[i]]
		if !present {
			errorString := fmt.Sprintf("%s is not valid", v)
			return nil, errors.New(errorString)
		}
	}
	return outs, nil
}

// DomainFrontedCollector is a {"domain": "a", "front": "b"} map
type DomainFrontedCollector struct {
	Domain string `json:"domain"`
	Front  string `json:"front"`
}

// Collectors holds the urls of onion, https, and domain-fronted collectors
type Collectors struct {
	Onion         []string                 `json:"onion"`
	HTTPS         []string                 `json:"https"`
	DomainFronted []DomainFrontedCollector `json:"domain_fronted"`
}

// GetCollectors returns a map of collectors keyed by their type
func GetCollectors(db *sqlx.DB) (Collectors, error) {
	var (
		collectors Collectors
		err        error
	)
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
			continue
		}
		switch ctype {
		case "onion":
			collectors.Onion = append(collectors.Onion, caddress)
		case "https":
			collectors.HTTPS = append(collectors.HTTPS, caddress)
		case "domain_fronted":
			if !cfront.Valid {
				ctx.Error("domain_fronted collector with bad front domain")
				continue
			}
			collectors.DomainFronted = append(collectors.DomainFronted,
				DomainFrontedCollector{caddress, cfront.String})
		default:
			ctx.Error("collector with bad type in DB")
			continue
		}
	}
	return collectors, nil
}

// GetTestHelpers returns a map of test helpers keyed by the test name
func GetTestHelpers(db *sqlx.DB) (map[string][]string, error) {
	var (
		err error
	)
	helpers := make(map[string][]string)
	query := fmt.Sprintf(`SELECT
		test_name,
		address
		FROM %s`,
		pq.QuoteIdentifier(common.TestHelpersTable))
	rows, err := db.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return helpers, nil
		}
		ctx.WithError(err).Error("failed to get test helpers")
		return helpers, err
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
		helpers[testName] = append(helpers[testName], address)
	}
	return helpers, nil
}

// buildTestInputQuery returns the query string to get all the inputs for the
// given countries and category codes
func buildTestInputQuery(countries []string, catCodes []string) string {
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
	if len(catCodes) > 0 {
		query += " AND cat_code = ANY($3)"
	}
	query += " ORDER BY random() LIMIT $1"
	return query
}

// GetTestInputs returns a slice of test inputs
func GetTestInputs(countries []string, catCodes []string, count int64, db *sqlx.DB) ([]map[string]string, error) {
	var (
		err error
	)
	inputs := make([]map[string]string, 0)
	query := buildTestInputQuery(countries, catCodes)
	args := []interface{}{count, pq.StringArray(countries)}
	if len(catCodes) > 0 {
		args = append(args, pq.StringArray(catCodes))
	}

	stmt, err := db.Prepare(query)
	if err != nil {
		ctx.WithError(err).Error("failed to prepare query")
		return inputs, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Debugf("got an empty result")
			return inputs, nil
		}
		ctx.WithError(err).Error("failed to get test inputs (urls)")
		return inputs, err
	}
	for rows.Next() {
		var (
			url    string
			cat    string
			alpha2 string
		)
		err = rows.Scan(&url, &cat, &alpha2)
		if err != nil {
			ctx.WithError(err).Error("failed to get test input row (urls)")
			continue
		}
		input := map[string]string{"cat_code": cat, "url": url, "country": alpha2}
		inputs = append(inputs, input)
	}
	return inputs, nil
}

// RendezvousHandler handler for /rendezvous
func RendezvousHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	collectors, err := GetCollectors(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}
	testHelpers, err := GetTestHelpers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}

	// We use XX to denote ANY country
	probeCc := c.Query("probe_cc")
	countries := []string{"XX"}
	if probeCc != "" {
		countries = append(countries, probeCc)
	}
	countriesUpper, err := upperAndWhitelist(countries, allCountryCodes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	catParam := c.Query("cat_code")
	cats := []string{}
	if catParam != "" {
		cats = strings.Split(catParam, ",")
	}
	catsUpper, err := upperAndWhitelist(cats, allCatCodes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	countString := c.DefaultQuery("count",
		viper.GetString("api.default-inputs-to-return"))
	var count int64
	count, err = strconv.ParseInt(countString, 10, 64)
	if err != nil || count < 1 || count > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad count"})
		return
	}
	testInputs, err := GetTestInputs(countriesUpper, catsUpper, count, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "server side error"})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"collectors": collectors,
			"test_helpers": testHelpers,
			"inputs":       testInputs})
	return
}
