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
	uuid "github.com/satori/go.uuid"
)

// ExperimentData is the data for the task
type ExperimentData struct {
	ID               string `json:"id"`
	ExperimentNo     int64  `json:"-"`
	TestName         string `json:"test_name" binding:"required"`
	SigningKeyID     string `json:"signing_key_id"`
	SignedExperiment string `json:"signed_experiment"`
	ArgsIdx          []int  `json:"args_idx"`
	State            string `json:"state"`
}

var validSigningKeys = map[string]*rsa.PublicKey{}

func loadSigningKeys() error {
	// ID: 581ec75b81726d2a0e8268ee0612531cc117e4302856e049f909d71bc8e42299
	keyPEM := []byte(`-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxfU1kBg7LwMmFR2DsObh
b6wL4fRfxOgSeXjcwYUg6LhF3yVVDyRLPMg0KUQoUlO+mscsLoiW6T02RFQgH2Y4
PjNt3XvpJwjGvLH4+qiB7rcJqRlkqdIVzonK1TOqBlspNAdj+SYeluj6+Z1mVisb
yVmUv8KIPLfPp4y2yPfdCEb/vZNck4VviWsjYPMO3RUV8hbnYqOC8XX1jEA84B73
xwuapz6PIP0EP02OvzO/g2ggOsaJjfGtc04OxnrXYLh6SAThQOdas4m3vXuooMsI
IqsOXuKwezyr5JQBDTuZL0uv4/X6iBD4mWWXGbg0vVmGVJttRJHL1IEJj2kDi3UW
ZwIDAQAB
-----END RSA PUBLIC KEY-----`)
	h := sha256.New()
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyPEM)
	if err != nil {
		return err
	}
	h.Write(keyPEM)
	// It's a bit weird to use the ascii PEM encoding of the key to do a ID, but
	// for the moment it's as good as anything.
	keyID := hex.EncodeToString(h.Sum(nil))
	validSigningKeys[keyID] = pubKey
	return nil
}

// ParseSignedExperiment reads a JWT token of a signed experiment
func ParseSignedExperiment(ed *ExperimentData) (*jwt.Token, error) {
	verifyKey, ok := validSigningKeys[ed.SigningKeyID]
	if !ok {
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

// CreateExperimentForClient creates a new task and stores it in the JobDB
func (j *Job) CreateExperimentForClient(jDB *JobDB, cID string, ed *ExperimentData) error {
	tx, err := jDB.db.Begin()
	if err != nil {
		ctx.WithError(err).Error("failed to open CreateExperimentForClient transaction")
		return err
	}

	ed.ID = uuid.NewV4().String()
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
			return err
		}
		defer stmt.Close()

		token, err := ParseSignedExperiment(ed)
		if err != nil {
			ctx.WithError(err).Error("failed to ParseSignedExperiment")
			return err
		}
		var argsIdx []int
		args := token.Claims.(jwt.MapClaims)["args"].([]interface{})
		// We just add all the indexes for the moment
		for i := 0; i <= len(args); i++ {
			argsIdx = append(argsIdx, i)
		}

		now := time.Now().UTC()
		_, err = stmt.Exec(ed.ID, cID,
			ed.ExperimentNo, pq.Array(argsIdx),
			"ready", 0,
			now, nil,
			nil, nil, now)
		if err != nil {
			tx.Rollback()
			ctx.WithError(err).Error("failed to insert into tasks table")
			return err
		}
		if err = tx.Commit(); err != nil {
			ctx.WithError(err).Error("failed to commit transaction in tasks table, rolling back")
			return err
		}
	}

	return nil
}

// NewExperimentData creates a new ExperimentData struct
func NewExperimentData(jDB *JobDB, expNo int64) (*ExperimentData, error) {
	ed := ExperimentData{ExperimentNo: expNo}
	query := `SELECT
			id,
			test_name,
			signed_experiment
			FROM experiment_jobs
			WHERE experiment_no = $1`

	err := jDB.db.QueryRow(query, expNo).Scan(
		&ed.ID,
		&ed.TestName,
		&ed.SignedExperiment)
	if err != nil {
		ctx.WithError(err).Errorf("failed to get experiment_no %d", expNo)
		return nil, err
	}
	return &ed, nil
}

// GetExperiment returns the experiment specfic to a certain user
func GetExperiment(db *sqlx.DB, experimentID string) (ExperimentData, string, error) {
	var (
		err     error
		probeID string
	)
	exp := ExperimentData{}
	query := `SELECT
		client_experiments.id, client_experiments.probe_id,
		client_experiments.experiment_no, client_experiments.args_idx,
		client_experiments.state,
		job_experiments.test_name, job_experiments.signing_key_id,
		job_experiments.signed_experiment
		FROM client_experiments
		WHERE client_experiments.id = $1
		JOIN job_experiments
		ON job_experiments.experiment_no = client_experiments.experiment_no`
	err = db.QueryRow(query, experimentID).Scan(
		&exp.ID, &probeID,
		&exp.ExperimentNo, pq.Array(&exp.ArgsIdx),
		&exp.State,
		&exp.TestName, &exp.SigningKeyID,
		&exp.SignedExperiment)
	if err != nil {
		if err == sql.ErrNoRows {
			return exp, probeID, ErrTaskNotFound
		}
		ctx.WithError(err).Error("failed to get task")
		return exp, probeID, err
	}

	return exp, probeID, nil
}

// SetExperimentState sets the state of the task
func SetExperimentState(expID string, uID string,
	state string, validStates []string,
	updateTimeCol string,
	db *sqlx.DB) error {
	var err error
	experimentData, probeID, err := GetExperiment(db, expID)
	if probeID != uID {
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
func SetExperimentNotified(jDB *JobDB, jt *JobTarget) error {
	return SetExperimentState(
		jt.ExperimentData.ID,
		jt.ClientID,
		"notified",
		[]string{"ready"},
		"notification_time",
		jDB.db)
}
