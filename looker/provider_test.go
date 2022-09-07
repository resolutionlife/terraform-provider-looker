package looker

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/looker-open-source/sdk-codegen/go/rtl"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

var (
	testAccProviders map[string]func() (*schema.Provider, error)
	testAccProvider  *schema.Provider
)

func init() {
	testAccProvider = NewProvider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"looker": func() (*schema.Provider, error) { return testAccProvider, nil },
	}
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
