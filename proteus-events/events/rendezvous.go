package events

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/spf13/viper"
)

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
		pq.QuoteIdentifier(viper.GetString("database.collectors-table")))
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
		pq.QuoteIdentifier(viper.GetString("database.test-helpers-table")))
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
		pq.QuoteIdentifier(viper.GetString("database.urls-table")),
		pq.QuoteIdentifier(viper.GetString("database.countries-table")),
		pq.QuoteIdentifier(viper.GetString("database.url-categories-table")))
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
