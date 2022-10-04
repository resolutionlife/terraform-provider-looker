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

// NewProvider returns an new provider.
func NewProvider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"base_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_secret": {
				Type:     schema.TypeString,
				Required: true,
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
			"looker_role":           datasourceRole(),
			"looker_group":          datasourceGroup(),
			"looker_model_set":      datasourceModelSet(),
			"looker_permission_set": dataSourcePermissionSet(),
			"looker_idp_metadata":   datasourceLookerIdpMetadata(),
		},
		ConfigureContextFunc: configureProvider,
	}

	return provider
}

// configureProvider uses the environment variables to create a Looker client.
func configureProvider(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	authSession := rtl.NewAuthSession(rtl.ApiSettings{
		BaseUrl:      data.Get("base_url").(string),
		ClientId:     data.Get("client_id").(string),
		ClientSecret: data.Get("client_secret").(string),
		ApiVersion:   "4.0", // this provider only supports API verison 4.0
		VerifySsl:    data.Get("verify_ssl").(bool),
		Timeout:      int32(data.Get("timeout").(int)),
		AgentTag:     fmt.Sprintf("Terraform Looker Provider (%s)", version.ProviderVersion),
	})

	return client.NewLookerSDK(authSession), nil
}
