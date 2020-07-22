package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/common"
	"github.com/spf13/viper"
)

// prepareURLsQuery returns the statement to get all the inputs for the
// given countries and category codes
func prepareURLsQuery(q URLsQuery, db *sqlx.DB) (*sql.Stmt, []interface{}, error) {
	var (
		countryCodes []string
		args         []interface{}
	)
	markerIdx := 0
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
		INNER JOIN %s url_cats ON urls.cat_no = url_cats.cat_no
		WHERE active = true`,
		pq.QuoteIdentifier(common.URLsTable),
		pq.QuoteIdentifier(common.CountriesTable),
		pq.QuoteIdentifier(common.URLCategoriesTable))
	// countries is always greater than zero
	markerIdx++
	query += fmt.Sprintf(" AND alpha_2 = ANY($%d)", markerIdx)
	args = append(args, pq.StringArray(countryCodes))
	if q.CategoryCodes != "" {
		markerIdx++
		query += fmt.Sprintf(" AND cat_code = ANY($%d)", markerIdx)
		args = append(args, pq.StringArray(
			common.MapToUppercase(strings.Split(q.CategoryCodes, ","))))
	}
	query += " ORDER BY random()"
	if q.Limit > 0 {
		args = append(args, q.Limit)
		markerIdx++
		query += fmt.Sprintf(" LIMIT $%d", markerIdx)
	}
	stmt, err := db.Prepare(query)
	return stmt, args, err
}

// URLInfo holds the name, type and address of a test helper
type URLInfo struct {
	CategoryCode string `json:"category_code"`
	URL          string `json:"url"`
	CountryCode  string `json:"country_code"`
}

func isValidURL(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		// XXX maybe this should be a more serious error
		ctx.WithError(err).Errorf("%s url is invalid", urlStr)
		return false
	}
	if u.Path == "" {
		ctx.Errorf("%s url contains empty path", urlStr)
		return false
	}
	if u.Host == "" {
		ctx.Errorf("%s url contains empty host", urlStr)
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		ctx.Errorf("%s url scheme is not http or https", urlStr)
		return false
	}
	return true
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
		var ui URLInfo
		err = rows.Scan(&ui.URL, &ui.CategoryCode, &ui.CountryCode)
		if err != nil {
			ctx.WithError(err).Error("failed to get test input row (urls)")
			continue
		}
		if isValidURL(ui.URL) != true {
			ctx.Errorf("%s invalid URL skipping", ui.URL)
			continue
		}
		urls = append(urls, ui)
	}
	return urls, nil
}

// URLsQuery is the user issued request for URLs
type URLsQuery struct {
	Limit         int64  `form:"limit"`
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

// URLsHandler returns the list of requested URLs
func URLsHandler(c *gin.Context) {
	var (
		err       error
		urlsQuery URLsQuery
	)
	// This is equivalent to setting no limit
	urlsQuery.Limit = -1

	if common.ValidateCSVMapStr(urlsQuery.CountryCode, common.AllCountryCodes) == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country_code"})
		return
	}
	if common.ValidateCSVMapStr(urlsQuery.CategoryCodes, common.AllCategoryCodes) == false {
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
	if urlsQuery.Limit < 0 {
		metadata["count"] = len(urls)
		metadata["pages"] = 1
	}
	c.JSON(http.StatusOK,
		gin.H{
			"metadata": metadata,
			"results":  urls,
		})
	return
}

// PsiphonConfigHandler returns the psiphon configuration.
func PsiphonConfigHandler(c *gin.Context) {
	content, err := ioutil.ReadFile(viper.GetString("psiphon.config-file"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server side error",
		})
		return
	}
	c.Data(http.StatusOK, "application/json", content)
}

// BridgeInfo is the metadata of a tor bridge
type BridgeInfo struct {
	Address     string                 `json:"address"`
	Fingerprint string                 `json:"fingerprint,omitempty"`
	Port        int                    `json:"port"`
	Protocol    string                 `json:"protocol"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// BridgeMap is a mapping between the bridge ID (to be shown in the OONI Probe
// UI) and the BridgeInfo
type BridgeMap map[string]BridgeInfo

func lookupPrivateBridges(countryCode string) (BridgeMap, error) {
	var reqURL = fmt.Sprintf("https://bridges.torproject.org/wolpertinger/bridges?id=&type=ooni&country_code=%s", countryCode)
	bridgeMap := BridgeMap{}

	client := http.DefaultClient
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return bridgeMap, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", viper.GetString("tor.bridges-api-key")))
	resp, err := client.Do(req)
	if err != nil {
		return bridgeMap, err
	}
	defer resp.Body.Close()
	log.Debugf("GET %s", reqURL)
	log.Debugf("%d %s", resp.StatusCode, resp.Status)
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Errorf("Got bad response %s %d %s", reqURL, resp.StatusCode, body)
		return bridgeMap, errors.New("Bad response")
	}
	if err := json.NewDecoder(resp.Body).Decode(&bridgeMap); err != nil {
		log.Errorf("Got decoding error: %s %d %+v", reqURL, resp.StatusCode, err)
		return bridgeMap, err
	}
	log.Debugf("bridgeMap %v", bridgeMap)
	return bridgeMap, nil
}

// TorTargetsHandler returns the targets for the tor nettest.
func TorTargetsHandler(c *gin.Context) {
	countryCode := c.Query("country_code")

	finalBridgeMap := BridgeMap{}
	content, err := ioutil.ReadFile(viper.GetString("tor.targets-file"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server side error",
		})
		return
	}
	err = json.Unmarshal(content, &finalBridgeMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server side error",
		})
		// If we don't have anything meaningful to return to the user
		// from our default list, then it's a hard error.
		return
	}
	if (len(viper.GetString("tor.bridges-api-key")) > 0) && countryCode != "" {
		log.Debug("Requesting bridges from bridgedb")
		tpoBridgeMap, err := lookupPrivateBridges(countryCode)
		if err != nil {
			// We don't explicitly handle this error. Here's why:
			//
			// Be flexible here and return _some_ information to the user
			// rather than a hard error. Ideally, here we would like to have
			// a prometheus metric counting the number of times in which
			// this specific query is failing. For now, we just log inside
			// of the lookupPrivateBridged function.
		}
		for k, v := range tpoBridgeMap {
			v.Source = "bridgedb"
			finalBridgeMap[k] = v
		}
	}
	c.JSON(http.StatusOK, finalBridgeMap)
}
