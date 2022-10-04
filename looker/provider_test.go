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
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/looker-open-source/sdk-codegen/go/rtl"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

var (
	testAccProviders map[string]func() (*schema.Provider, error)
	testAccProvider  *schema.Provider
)

func NewTestProvider(cassettePath string) func() error {
	recMode := os.Getenv("TF_REC")

	rec := recorder.ModeReplayOnly
	if recMode != "" {
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
	r.AddHook(filterAuthHeaders, recorder.AfterCaptureHook)
	r.AddHook(filterCredentials, recorder.BeforeSaveHook)

	r.SetMatcher(customBodyMatcher)

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

func redactBody(body *string, filterKeys []string) error {
	var responseBody map[string]interface{}
	err := json.Unmarshal([]byte(*body), &responseBody)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal body")
	}

	for _, key := range filterKeys {
		_, ok := responseBody[key]
		if ok {
			responseBody[key] = "[REDACTED]"
		}
	}

	b, err := json.Marshal(responseBody)
	if err != nil {
		return errors.Wrap(err, "failed to marshal body")
	}

	*body = string(b)

	return nil
}

func filterCredentials(i *cassette.Interaction) error {
	if strings.Contains(i.Request.Body, "client_id") ||
		strings.Contains(i.Request.Body, "client_secret") {
		err := redactBody(&i.Response.Body, []string{"access_token"})
		if err != nil {
			return errors.Wrap(err, "failed to redact response body")
		}
	}

	if strings.Contains(i.Response.Body, "client_id") ||
		strings.Contains(i.Response.Body, "client_secret") {
		err := redactBody(&i.Response.Body, []string{"client_id", "client_secret"})
		if err != nil {
			return errors.Wrap(err, "failed to redact response body")
		}
	}

	_, ok := i.Request.Form["client_id"]
	if ok {
		i.Request.Form.Set("client_id", "[REDACTED]")
	}
	_, ok = i.Request.Form["client_secret"]
	if ok {
		i.Request.Form.Set("client_secret", "[REDACTED]")
	}

	requestUrl, err := url.Parse(i.Request.URL)
	if err != nil {
		return err
	}

	if path.Base(requestUrl.Path) == "login" {
		i.Request.Body = "[REDACTED]"
	}

	return nil
}

func filterAuthHeaders(i *cassette.Interaction) error {
	delete(i.Request.Headers, "Authorization")
	return nil
}

func customBodyMatcher(r *http.Request, i cassette.Request) bool {
	if i.Body == "[REDACTED]" {
		return true
	}

	if r.Body == nil || r.Body == http.NoBody {
		return cassette.DefaultMatcher(r, i)
	}

	var reqBody []byte
	var err error
	reqBody, err = io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("failed to read request body")
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	return r.Method == i.Method && r.URL.String() == i.URL && string(reqBody) == i.Body
}
