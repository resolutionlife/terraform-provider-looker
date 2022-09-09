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
			"looker_role":                resourceRole(),
			"looker_user":                resourceUser(),
			"looker_group":               resourceGroup(),
			"looker_user_roles":          resourceUserRoles(),
			"looker_permission_set":      resourcePermissionSet(),
			"looker_model_set":           resourceModelSet(),
			"looker_group_user":          resourceGroupUser(),
			"looker_user_attribute":      resourceUserAttribute(),
			"looker_group_group":         resourceGroupGroup(),
			"looker_role_groups":         resourceRoleGroups(),
			"looker_user_attribute_user": resourceUserAttributeUser(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"looker_role":           datasourceRole(),
			"looker_model_set":      datasourceModelSet(),
			"looker_permission_set": dataSourcePermissionSet(),
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
