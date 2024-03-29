package looker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/looker-open-source/sdk-codegen/go/rtl"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const dummyAPIHostname = "example.cloud.looker.com"

var (
	testAccProviders map[string]func() (*schema.Provider, error)
	testAccProvider  *schema.Provider
)

func NewTestProvider(cassettePath string) func() error {
	recMode := os.Getenv("TF_REC")

	rec := recorder.ModeReplayOnly
	if recMode == "1" {
		rec = recorder.ModeRecordOnly
	}

	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               rec,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	})
	if err != nil {
		log.Fatalf("failed to create new recorder: %v", err)
	}

	// ensures creds do not leak
	r.AddHook(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}, recorder.AfterCaptureHook)
	r.AddHook(filterCredentials, recorder.BeforeSaveHook)

	r.SetMatcher(customMatcher)

	testAccProvider = NewProvider(WithRecorder(r))
	testAccProviders = map[string]func() (*schema.Provider, error){
		"looker": func() (*schema.Provider, error) { return testAccProvider, nil },
	}

	return r.Stop
}

func newTestLookerSDK() (*client.LookerSDK, error) {
	apiSettings, err := rtl.NewSettingsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("unable to create Looker client settings from environment variables: %w", err)
	}
	return client.NewLookerSDK(rtl.NewAuthSession(apiSettings)), nil
}

// TestMain is used to parse special test flags to invoke sweepers.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func redactJSON(body *string, filterKeys []string) error {
	var responseBody map[string]interface{}
	err := json.Unmarshal([]byte(*body), &responseBody)
	if err != nil {
		return fmt.Errorf("failed to unmarshal body: %w", err)
	}

	for _, key := range filterKeys {
		_, ok := responseBody[key]
		if ok {
			responseBody[key] = "[REDACTED]"
		}
	}

	b, err := json.Marshal(responseBody)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	*body = string(b)

	return nil
}

func filterCredentials(i *cassette.Interaction) error {
	if strings.Contains(i.Request.Headers.Get("Content-Type"), "application/json") {
		redactJSON(&i.Request.Body, []string{"access_token", "client_id", "client_secret"}) //nolint:errcheck
	}

	if strings.Contains(i.Response.Headers.Get("Content-Type"), "application/json") {
		redactJSON(&i.Response.Body, []string{"access_token", "client_id", "client_secret"}) //nolint:errcheck
	}

	_, ok := i.Request.Form["client_id"]
	if ok {
		i.Request.Form.Set("client_id", "[REDACTED]")
	}
	_, ok = i.Request.Form["client_secret"]
	if ok {
		i.Request.Form.Set("client_secret", "[REDACTED]")
	}

	requestURL, err := url.Parse(i.Request.URL)
	if err != nil {
		return err
	}

	if path.Base(requestURL.Path) == "login" {
		i.Request.Body = "[REDACTED]"
	}

	i.Request.Host = dummyAPIHostname
	requestURL.Host = dummyAPIHostname
	i.Request.URL = requestURL.String()

	return nil
}

func customMatcher(r *http.Request, i cassette.Request) bool {
	u, parseErr := url.Parse(i.URL)
	if parseErr != nil {
		log.Fatal("failed to parse request url")
	}

	if r.Method != i.Method || r.URL.Path != u.Path || !reflect.DeepEqual(r.URL.Query(), u.Query()) {
		return false
	}

	if r.Body == nil || r.Body == http.NoBody || i.Body == "[REDACTED]" {
		return true
	}

	reqBody, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		log.Fatal("failed to read request body")
	}

	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var req, cassette interface{}
		err := json.Unmarshal(reqBody, &req)
		if err != nil {
			return false
		}

		err = json.Unmarshal([]byte(i.Body), &cassette)
		if err != nil {
			return false
		}

		return reflect.DeepEqual(req, cassette)
	}

	return string(reqBody) == i.Body
}
