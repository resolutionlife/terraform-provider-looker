package looker

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/looker-open-source/sdk-codegen/go/rtl"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

var (
	testAccProviders map[string]func() (*schema.Provider, error)
	testAccProvider  *schema.Provider
)

func NewTestProvider(cassettePath string) func() error {
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               recorder.ModeRecordOnce,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	})
	if err != nil {
		log.Fatalf("failed to create new recorder: %v", err)
	}

	// ensures creds do not leak
	r.AddHook(filterAuthHeaders, recorder.AfterCaptureHook)
	r.AddHook(filterCredentials, recorder.BeforeSaveHook)

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

func filterCredentials(i *cassette.Interaction) error {
	if strings.Contains(i.Request.Body, "client_id") || strings.Contains(i.Request.Body, "client_secret") {
		form := make(url.Values)
		form.Add("client_id", "[REDACTED]")
		form.Add("client_secret", "[REDACTED]")

		i.Request.Form = form
		i.Request.Body = "[REDACTED]"
		i.Response.Body = `{"access_token": "[REDACTED]"}`
	}

	return nil
}

func filterAuthHeaders(i *cassette.Interaction) error {
	delete(i.Request.Headers, "Authorization")
	return nil
}
