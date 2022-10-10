package looker

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/version"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/looker-open-source/sdk-codegen/go/rtl"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

type ProviderOptions func(*schema.Provider)

func WithRecorder(rec *recorder.Recorder) ProviderOptions {
	return func(p *schema.Provider) {
		p.ConfigureContextFunc = configWrapper(rec)
	}
}

// NewProvider returns an new provider.
func NewProvider(opts ...ProviderOptions) *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKERSDK_BASE_URL", nil),
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKERSDK_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKERSDK_CLIENT_SECRET", nil),
			},
			"verify_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKERSDK_VERIFY_SSL", true),
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LOOKERSDK_TIMEOUT", 120),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"looker_role":                 resourceRole(),
			"looker_user":                 resourceUser(),
			"looker_group":                resourceGroup(),
			"looker_user_roles":           resourceUserRoles(),
			"looker_permission_set":       resourcePermissionSet(),
			"looker_model_set":            resourceModelSet(),
			"looker_group_user":           resourceGroupUser(),
			"looker_group_group":          resourceGroupGroup(),
			"looker_role_groups":          resourceRoleGroups(),
			"looker_user_attribute":       resourceUserAttribute(),
			"looker_user_attribute_user":  resourceUserAttributeUser(),
			"looker_user_attribute_group": resourceUserAttributeGroup(),
			"looker_user_api_client":      resourceUserAPIClient(),
			"looker_saml_config":          resourceSamlConfig(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"looker_role":           dataSourceRole(),
			"looker_group":          dataSourceGroup(),
			"looker_model_set":      dataSourceModelSet(),
			"looker_permission_set": dataSourcePermissionSet(),
			"looker_idp_metadata":   dataSourceLookerIdpMetadata(),
		},
		ConfigureContextFunc: configWrapper(nil),
	}

	for _, opt := range opts {
		opt(provider)
	}

	return provider
}

func configWrapper(rec *recorder.Recorder) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiSettings := rtl.ApiSettings{
			BaseUrl:      d.Get("base_url").(string),
			ClientId:     d.Get("client_id").(string),
			ClientSecret: d.Get("client_secret").(string),
			ApiVersion:   "4.0", // this provider only supports API version 4.0
			VerifySsl:    d.Get("verify_ssl").(bool),
			Timeout:      int32(d.Get("timeout").(int)),
			AgentTag:     fmt.Sprintf("Terraform Looker Provider (%s)", version.ProviderVersion),
		}

		var authSession *rtl.AuthSession
		if rec == nil {
			authSession = rtl.NewAuthSession(apiSettings)
		} else {
			authSession = rtl.NewAuthSessionWithTransport(apiSettings, rec)
		}

		return client.NewLookerSDK(authSession), nil
	}
}
