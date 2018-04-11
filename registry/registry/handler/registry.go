package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	common "github.com/ooni/orchestra/common"
	"golang.org/x/crypto/bcrypt"
)

var ctx = log.WithFields(log.Fields{
	"pkg": "handler",
	"cmd": "ooni-registry",
})

// ClientData metadata about a client
type ClientData struct {
	ProbeCC  string `json:"probe_cc" binding:"required"`
	ProbeASN string `json:"probe_asn" binding:"required"`
	Platform string `json:"platform" binding:"required"`

	SoftwareName    string   `json:"software_name" binding:"required"`
	SoftwareVersion string   `json:"software_version" binding:"required"`
	SupportedTests  []string `json:"supported_tests" binding:"required"`

	NetworkType        string `json:"network_type"`
	AvailableBandwidth string `json:"available_bandwidth"`
	Language           string `json:"language"`

	Token string `json:"token"`

	ProbeFamily string `json:"probe_family"`
	ProbeID     string `json:"probe_id"`

	Password string `json:"password"`
}

// IsClientRegistered checks is a client is registered
func IsClientRegistered(db *sqlx.DB, clientID string) (bool, error) {
	var found string
	query := fmt.Sprintf(`SELECT id FROM %s WHERE id = $1`,
		pq.QuoteIdentifier(common.ActiveProbesTable))
	err := db.QueryRow(query, clientID).Scan(&found)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Update the metadata for a client
func Update(db *sqlx.DB, clientID string, req ClientData) error {
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
			lang_code,
			token, probe_family,
			probe_id, update_type
		) VALUES (
			$1, $2,
			$3,
			$4, $5,
			$6, $7,
			$8, $9,
			$10, $11,
			$12,
			$13, $14,
			$15, $16)`,
			pq.QuoteIdentifier(common.ProbeUpdatesTable))

		stmt, err := tx.Prepare(query)
		if err != nil {
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
			req.Language,
			req.Token, req.ProbeFamily,
			req.ProbeID, "register")
		if err != nil {
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
			lang_code = $11,
			token = $12,
			probe_family = $13,
			probe_id = $14
			WHERE id = $1`,
			pq.QuoteIdentifier(common.ActiveProbesTable))

		stmt, err := tx.Prepare(query)
		if err != nil {
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
			req.Language,
			req.Token,
			req.ProbeFamily,
			req.ProbeID)
		if err != nil {
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

// Register a new client
func Register(db *sqlx.DB, req ClientData) (string, error) {
	if (req.Platform == "ios" || req.Platform == "android") && req.Token == "" {
		return "", errors.New("missing device token")
	}
	if req.Password == "" {
		return "", errors.New("missing password")
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
			lang_code,
			token, probe_family,
			probe_id
		) VALUES (
			$1, $2,
			$3,
			$4, $5,
			$6, $7,
			$8, $9,
			$10, $11,
			$12,
			$13, $14,
			$15)`,
			pq.QuoteIdentifier(common.ActiveProbesTable))

		stmt, err := tx.Prepare(query)
		if err != nil {
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
			req.Language,
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
			lang_code,
			token, probe_family,
			probe_id, update_type
		) VALUES (
			$1, $2,
			$3,
			$4, $5,
			$6, $7,
			$8, $9,
			$10, $11,
			$12,
			$13, $14,
			$15, $16)`,
			pq.QuoteIdentifier(common.ProbeUpdatesTable))

		stmt, err := tx.Prepare(query)
		if err != nil {
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
			req.Language,
			req.Token, req.ProbeFamily,
			req.ProbeID, "register")
		if err != nil {
			ctx.WithError(err).Error("failed to add data to update table, rolling back")
			tx.Rollback()
			return "", errors.New("error in adding data to update probes")
		}
	}

	{
		query := fmt.Sprintf(`INSERT INTO %s (
			username,
			password_hash,
			last_access,
			role
		) VALUES (
			$1,
			$2,
			$3,
			$4)`,
			pq.QuoteIdentifier(common.AccountsTable))

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.WithError(err).Error("failed to hash password")
			return "", err
		}

		stmt, err := tx.Prepare(query)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare accounts query")
			return "", err
		}
		defer stmt.Close()

		_, err = stmt.Exec(clientID,
			string(passwordHash),
			time.Now().UTC(),
			"device")
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into accounts table, rolling back")
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		ctx.WithError(err).Error("failed to commit transaction, rolling back")
		return "", err
	}

	return clientID, nil
}

// UpdateHandler device endpoint for registered probes to update their metadata
func UpdateHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	var updateReq ClientData
	clientID := c.Param("client_id")
	err := c.BindJSON(&updateReq)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}
	isRegistered, err := IsClientRegistered(db, clientID)
	if err != nil {
		ctx.WithError(err).Error("failed to learn if client is registered")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	if isRegistered == false {
		c.JSON(http.StatusNotFound,
			gin.H{"error": "client is not registered"})
		return
	}

	err = Update(db, clientID, updateReq)
	if err != nil {
		ctx.WithError(err).Error("failed to update")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK,
		gin.H{"status": "ok"})
}

// RegisterHandler this is the public endpoint handler for registering new clients
func RegisterHandler(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)

	var registerReq ClientData
	err := c.BindJSON(&registerReq)
	if err != nil {
		ctx.WithError(err).Error("invalid request")
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "invalid request"})
		return
	}

	clientID, err := Register(db, registerReq)
	if err != nil {
		c.JSON(http.StatusBadRequest,
			gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client_id": clientID})
	return
}
