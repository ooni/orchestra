package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/ooni/orchestra/registry/registry/handler"
)

var registryBaseURL = "http://127.0.0.1:8080"
var orchestrateBaseURL = "http://127.0.0.1:8082"
var gorushBaseURL = "http://127.0.0.1:8081"

// os.Getenv("ORCHESTRATE_URL")
// os.Getenv("GORUSH_URL")
// os.Getenv("REGISTRY_URL")

func resolveBaseURL(baseURL string, path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(u).String(), nil
}

func doRequestJSON(baseURL string, path string, method string, reqJSON interface{}) (*map[string]interface{}, error) {
	body, err := json.Marshal(reqJSON)
	if err != nil {
		return nil, err
	}
	url, err := resolveBaseURL(baseURL, path)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	request, _ := http.NewRequest(
		method,
		url,
		bytes.NewReader(body),
	)
	resp, _ := client.Do(request)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func TestRegistry(t *testing.T) {
	reqJSON := handler.ClientData{
		ProbeCC:            "IT",
		ProbeASN:           "AS1234",
		Platform:           "android",
		SoftwareName:       "ooni-testing",
		SoftwareVersion:    "0.0.1",
		SupportedTests:     []string{"web_connectivity"},
		NetworkType:        "wifi",
		AvailableBandwidth: "100",
		Language:           "en",
		Token:              "XXXX-TESTING",
		Password:           "testing",
	}
	result, err := doRequestJSON(registryBaseURL, "/api/v1/register", "POST", reqJSON)
	if err != nil {
		fmt.Printf("error %v", err)
	}
	fmt.Printf("resp: %v", result)
}
