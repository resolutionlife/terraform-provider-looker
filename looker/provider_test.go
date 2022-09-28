package looker

import (
	"fmt"
	"log"
	"net/http"
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

// func init() {
// 	r, err := recorder.NewWithOptions(&recorder.Options{
// 		CassetteName:       cassettePath,
// 		Mode:               recorder.ModeRecordOnly,
// 		SkipRequestLatency: true,
// 		RealTransport:      http.DefaultTransport,
// 	})
// 	if err != nil {
// 		log.Fatalf("failed to create new recorder: %v", err)
// 	}

// 	// want to stop only at end of all tests
// 	defer r.Stop()

// 	if !r.IsRecording() {
// 		log.Fatalf("recorder not recording to cassette: %v", err)
// 	}

// 	r.AddHook(func(i *cassette.Interaction) error {
// 		fmt.Printf("Req: %v \n", i.Request.Body)
// 		fmt.Println("")
// 		fmt.Printf("Res: %v \n", i.Response.Body)
// 		fmt.Println("============================")
// 		return nil
// 	}, recorder.AfterCaptureHook)

// 	testAccProvider = NewProvider(WithRecorder(r))
// 	testAccProviders = map[string]func() (*schema.Provider, error){
// 		"looker": func() (*schema.Provider, error) { return testAccProvider, nil },
// 	}
// }

func NewTestProvider(cassettePath string) func() error {
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               recorder.ModeReplayWithNewEpisodes,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	})
	if err != nil {
		log.Fatalf("failed to create new recorder: %v", err)
	}

	// if !r.IsRecording() {
	// 	log.Fatalf("recorder not recording to cassette: %v", err)
	// }

	r.AddHook(func(i *cassette.Interaction) error {
		fmt.Printf("Req: %v \n", i.Request.Body)
		fmt.Println("")
		fmt.Printf("Res: %v \n", i.Response.Body)
		fmt.Println("============================")
		return nil
	}, recorder.AfterCaptureHook)

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
