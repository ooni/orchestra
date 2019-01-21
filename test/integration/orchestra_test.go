package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	jwt "github.com/hellais/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/ooni/orchestra/common/middleware"
	operator_cmd "github.com/ooni/orchestra/operator/cmd"
	orchestrate_cmd "github.com/ooni/orchestra/orchestrate/cmd"
	"github.com/ooni/orchestra/orchestrate/orchestrate/keystore"
	registry_handler "github.com/ooni/orchestra/registry/registry/handler"
)

const adminUsername = "test_admin"
const testingPassword = "testing"

var supportedTests = []string{
	"web_connectivity",
	"http_invalid_request_line",
	"http_header_field_manipulation",
	"ndt_test",
	"dash",
}

func newClientData(cc, token string) registry_handler.ClientData {
	return registry_handler.ClientData{
		ProbeCC:            cc,
		ProbeASN:           "AS1234",
		Platform:           "android",
		SoftwareName:       "ooni-testing",
		SoftwareVersion:    "0.0.1",
		SupportedTests:     supportedTests,
		NetworkType:        "wifi",
		AvailableBandwidth: "100",
		Language:           "en",
		Token:              token,
		Password:           testingPassword,
	}
}

func mapFromJSON(data []byte) map[string]interface{} {
	var result interface{}
	json.Unmarshal(data, &result)
	return result.(map[string]interface{})
}

func registerClient(r http.Handler, cd registry_handler.ClientData) (string, error) {
	w, err := performRequestJSON(r, "POST", "/api/v1/register", cd)
	if err != nil {
		return "", err
	}

	result := mapFromJSON(w.Body.Bytes())
	return result["client_id"].(string), nil
}

func updateClient(r http.Handler, clientID, authToken string, cd registry_handler.ClientData) (string, error) {
	w, err := performRequestJSONWithJWT(r, "PUT", "/api/v1/update/"+clientID, authToken, cd)
	if err != nil {
		return "", err
	}

	result := mapFromJSON(w.Body.Bytes())
	return result["status"].(string), nil
}

func login(r http.Handler, username, password string) (string, error) {
	reqJSON := middleware.Login{
		Username: username,
		Password: password,
	}
	w, err := performRequestJSON(r, "POST", "/api/v1/login", reqJSON)
	if err != nil {
		return "", err
	}

	result := mapFromJSON(w.Body.Bytes())
	return result["token"].(string), nil
}

func scheduleExperiment(r http.Handler, authToken, signedExperiment, comment, schedule string) (int64, error) {
	exp := map[string]interface{}{
		"comment":           comment,
		"delay":             0,
		"schedule":          schedule,
		"signed_experiment": signedExperiment,
		"target": map[string]interface{}{
			"countries": []string{},
			"platforms": []string{},
		},
	}

	w, err := performRequestJSONWithJWT(r, "POST", "/api/v1/admin/experiment", authToken, exp)
	if err != nil {
		return 1, err
	}

	result := mapFromJSON(w.Body.Bytes())
	return int64(result["id"].(float64)), nil
}

func TestAdminExperiment(t *testing.T) {
	err := orchTest.CleanDB()
	if err != nil {
		t.Fatal(err)
	}

	regRouter, err := NewRegistryRouter(orchTest.pgURL)
	if err != nil {
		t.Fatal(err)
	}

	orchRouter, err := NewOrchestrateRouter(orchTest.pgURL)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sqlx.Open("postgres", orchTest.pgURL)
	if err != nil {
		t.Fatal(err)
	}
	ud := orchestrate_cmd.UserData{
		Username: adminUsername,
		Password: testingPassword,
		KeyPath:  "testdata/ooni-orchestrate.pub",
	}
	err = orchestrate_cmd.AddUser(db, ud)
	if err != nil {
		t.Error(err)
	}

	token, err := login(regRouter, adminUsername, testingPassword)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Login token: %s\n", token)

	scheduleStr := "R2/2018-11-29T14:42:37.049Z/P7D"
	exp := keystore.OrchestraClaims{
		ProbeCC:  []string{},
		TestName: "web_connectivity",
		Schedule: scheduleStr,
		Args: map[string]interface{}{
			"urls": []map[string]string{
				{"url": "http://google.com", "code": "SRCH"},
			},
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    adminUsername,
		},
	}
	signedExperiment, err := operator_cmd.SignLocal("testdata/ooni-orchestrate.priv", exp)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Signed claims: %s\n", signedExperiment)

	expID, err := scheduleExperiment(orchRouter, token, signedExperiment, "web_connectivity test", scheduleStr)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Scheduled experiment: %d\n", expID)
}

func TestRegistryUpdate(t *testing.T) {
	err := orchTest.CleanDB()
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewRegistryRouter(orchTest.pgURL)
	if err != nil {
		t.Fatal(err)
	}

	cd := newClientData("IT", "XXX-TESTING")
	clientID, err := registerClient(r, cd)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Registered: %s\n", clientID)
	token, err := login(r, clientID, testingPassword)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Login token: %s\n", token)

	cd.ProbeCC = "GR"
	status, err := updateClient(r, clientID, token, cd)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Update status: %s\n", status)
}

func TestMain(m *testing.M) {
	orchTest = NewOrchestraTest()
	err := orchTest.Setup()
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	err = orchTest.Teardown()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(exitCode)
}
