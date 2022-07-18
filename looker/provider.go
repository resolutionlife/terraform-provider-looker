package looker

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/version"

	"github.com/looker-open-source/sdk-codegen/go/rtl"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

// Provider returns an resource provider for Looker
func Provider() *schema.Provider {
	provider := schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKER_BASE_URL", nil),
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKER_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKER_CLIENT_SECRET", nil),
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKER_VERSION", "4.0"),
			},
			"verify_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKER_VERIFY_SSL", true),
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKER_TIMEOUT", 120),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"looker_user": resourceUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}

	provider.ConfigureContextFunc = configureProvider

	return &provider
}

// configureProvider uses the environment variables to create a Looker client
func configureProvider(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	authSession := rtl.NewAuthSession(rtl.ApiSettings{
		BaseUrl:      data.Get("base_url").(string),
		ClientId:     data.Get("client_id").(string),
		ClientSecret: data.Get("client_secret").(string),
		ApiVersion:   data.Get("api_version").(string),
		VerifySsl:    data.Get("verify_ssl").(bool),
		Timeout:      int32(data.Get("timeout").(int)),
		AgentTag:     fmt.Sprintf("Terraform Looker Provider (%s)", version.ProviderVersion),
	})

	return client.NewLookerSDK(authSession), nil
}
