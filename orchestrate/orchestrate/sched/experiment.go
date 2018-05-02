package sched

import (
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	jwt "github.com/hellais/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ooni/orchestra/common"
	uuid "github.com/satori/go.uuid"
)

// ExperimentData is the data for the task
type ExperimentData struct {
	ExperimentNo     int64
	TestName         string
	SigningKeyID     string
	SignedExperiment string
	State            string
}

// ClientExperimentData is the data for the task
type ClientExperimentData struct {
	ID               string `json:"id"`
	ClientID         string `json:"client_id"`
	ExperimentNo     int64  `json:"-"`
	TestName         string `json:"test_name" binding:"required"`
	SigningKeyID     string `json:"signing_key_id"`
	SignedExperiment string `json:"signed_experiment"`
	ArgsIdx          []int  `json:"args_idx"`
	State            string `json:"state"`
}

var validSigningKeys = map[string]*rsa.PublicKey{}

func loadSigningKeys() error {
	// XXX this is just dummy testing key
	// Maybe we should move these keys into the database
	// ID: 581ec75b81726d2a0e8268ee0612531cc117e4302856e049f909d71bc8e42299
	keyPEM := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxfU1kBg7LwMmFR2DsObh
b6wL4fRfxOgSeXjcwYUg6LhF3yVVDyRLPMg0KUQoUlO+mscsLoiW6T02RFQgH2Y4
PjNt3XvpJwjGvLH4+qiB7rcJqRlkqdIVzonK1TOqBlspNAdj+SYeluj6+Z1mVisb
yVmUv8KIPLfPp4y2yPfdCEb/vZNck4VviWsjYPMO3RUV8hbnYqOC8XX1jEA84B73
xwuapz6PIP0EP02OvzO/g2ggOsaJjfGtc04OxnrXYLh6SAThQOdas4m3vXuooMsI
IqsOXuKwezyr5JQBDTuZL0uv4/X6iBD4mWWXGbg0vVmGVJttRJHL1IEJj2kDi3UW
ZwIDAQAB
-----END RSA PUBLIC KEY-----
`)
	h := sha256.New()
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyPEM)
	if err != nil {
		return err
	}
	h.Write(keyPEM)
	// It's a bit weird to use the ascii PEM encoding of the key to do a ID, but
	// for the moment it's as good as anything.
	keyID := hex.EncodeToString(h.Sum(nil))
	ctx.Debugf("adding to valid keys %s", validSigningKeys)
	validSigningKeys[keyID] = pubKey
	return nil
}

// ParseSignedExperiment reads a JWT token of a signed experiment
func ParseSignedExperiment(ed *ExperimentData) (*jwt.Token, error) {
	verifyKey, ok := validSigningKeys[ed.SigningKeyID]
	if !ok {
		ctx.Errorf("Did not find signing key: %s", ed.SigningKeyID)
		return nil, errors.New("Could not find signing key")
	}

	token, err := jwt.ParseWithClaims(ed.SignedExperiment, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return verifyKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// CreateClientExperiment creates a new experiment and adds it to the database
func CreateClientExperiment(db *sqlx.DB, ed *ExperimentData, cID string) (*ClientExperimentData, error) {
	// XXX maybe there is more powerful golang ideom for this
	clientExp := ClientExperimentData{
		ExperimentNo:     ed.ExperimentNo,
		ClientID:         cID,
		TestName:         ed.TestName,
		SignedExperiment: ed.SignedExperiment,
	}

	tx, err := db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open CreateExperimentForClient transaction")
		return nil, err
	}

	clientExp.ID = uuid.NewV4().String()
	{
		stmt, err := tx.Prepare(`INSERT INTO client_experiments (
			id, probe_id,
			experiment_no, args_idx,
			state, progress,
			creation_time, notification_time,
			accept_time, done_time,
			last_updated
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6,
			$7, $8,
			$9, $10,
			$11)`)
		if err != nil {
			ctx.WithError(err).Error("failed to prepare task create query")
			return nil, err
		}
		defer stmt.Close()

		token, err := ParseSignedExperiment(ed)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to ParseSignedExperiment")
			return nil, err
		}
		// XXX we may want to split this into some other function
		if ed.TestName == "web_connectivity" {
			urls := token.Claims.(jwt.MapClaims)["args"].(map[string]map[string]string)["urls"]
			// We just add all the indexes for the moment
			for i := 0; i <= len(urls); i++ {
				clientExp.ArgsIdx = append(clientExp.ArgsIdx, i)
			}
		}

		now := time.Now().UTC()
		_, err = stmt.Exec(clientExp.ID, clientExp.ClientID,
			clientExp.ExperimentNo, pq.Array(clientExp.ArgsIdx),
			"ready", 0,
			now, nil,
			nil, nil, now)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into tasks table")
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			ctx.WithError(err).Error("failed to commit transaction in tasks table, rolling back")
			return nil, err
		}
	}

	return &clientExp, nil
}

// NewExperimentData creates a new ExperimentData struct
func NewExperimentData(jDB *JobDB, expNo int64) (*ExperimentData, error) {
	ed := ExperimentData{ExperimentNo: expNo}
	query := fmt.Sprintf(`SELECT
			experiment_no,
			test_name,
			signed_experiment,
			signing_key_id
			FROM %s
			WHERE experiment_no = $1`,
		common.JobExperimentsTable)

	err := jDB.db.QueryRow(query, expNo).Scan(
		&ed.ExperimentNo,
		&ed.TestName,
		&ed.SignedExperiment,
		&ed.SigningKeyID)
	if err != nil {
		ctx.WithError(err).Errorf("failed to get experiment_no %d", expNo)
		return nil, err
	}
	return &ed, nil
}

// GetExperiment returns the experiment specfic to a certain user
func GetExperiment(db *sqlx.DB, experimentID string) (*ClientExperimentData, error) {
	var err error
	exp := ClientExperimentData{}
	query := fmt.Sprintf(`SELECT
		client_experiments.id, client_experiments.probe_id,
		client_experiments.experiment_no, client_experiments.args_idx,
		client_experiments.state,
		job_experiments.test_name, job_experiments.signing_key_id,
		job_experiments.signed_experiment
		FROM client_experiments
		JOIN job_experiments
		ON job_experiments.experiment_no = client_experiments.experiment_no
		WHERE client_experiments.id = $1`)
	err = db.QueryRow(query, experimentID).Scan(
		&exp.ID, &exp.ClientID,
		&exp.ExperimentNo, pq.Array(&exp.ArgsIdx),
		&exp.State,
		&exp.TestName, &exp.SigningKeyID,
		&exp.SignedExperiment)
	if err != nil {
		if err == sql.ErrNoRows {
			return &exp, ErrTaskNotFound
		}
		ctx.WithError(err).Error("failed to get task")
		return &exp, err
	}
	return &exp, nil
}

// SetExperimentState sets the state of the task
func SetExperimentState(expID string, uID string,
	state string, validStates []string,
	updateTimeCol string,
	db *sqlx.DB) error {
	var err error
	experimentData, err := GetExperiment(db, expID)
	if experimentData.ClientID != uID {
		return ErrAccessDenied
	}
	if err != nil {
		return err
	}
	stateConsistent := false
	for _, s := range validStates {
		if experimentData.State == s {
			stateConsistent = true
			break
		}
	}
	if !stateConsistent {
		return ErrInconsistentState
	}

	query := fmt.Sprintf(`UPDATE client_experiments SET
		state = $2,
		%s = $3,
		last_updated = $3
		WHERE id = $1`,
		updateTimeCol)

	_, err = db.Exec(query, expID, state, time.Now().UTC())
	if err != nil {
		ctx.WithError(err).Error("failed to get task")
		return err
	}
	return nil
}

// SetExperimentNotified marks the experiement of the JobTarget as notified
func SetExperimentNotified(jDB *JobDB, expID string, clientID string) error {
	return SetExperimentState(
		expID,
		clientID,
		"notified",
		[]string{"ready"},
		"notification_time",
		jDB.db)
}
