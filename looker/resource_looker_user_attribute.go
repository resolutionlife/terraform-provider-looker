package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceUserAttribute() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages Looker User Attributes",
		CreateContext: resourceUserAttributeCreate,
		ReadContext:   resourceUserAttributeRead,
		UpdateContext: resourceUserAttributeUpdate,
		DeleteContext: resourceUserAttributeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of the user attribute",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the user attribute, this is how you reference the attribute in Looker expressions and LookML",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user-friendly name displayed in the app for lists and filters",
			},
			"data_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"hidden": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "If set the value will be treated like a password, and once set, no one will be able to decrypt and view it",
			},
			"default_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value when no other value is set for the user or for one of the user's groups",
			},
			"user_can_view": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Required to use this attribute in query filters",
			},
			"user_can_edit": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Allows user to set their own value of this attribute",
			},
			"domain_whitelist": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "A list of urls that will be allowed as a destination for this user attribute, optionally using a wildcard '*'. You must set this when changing a user attribute to 'hidden'",
			},
		},
	}
}

func resourceUserAttributeCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttributes, err := api.CreateUserAttribute(
		sdk.WriteUserAttribute{
			Name:          *conv.PString(d.Get("name").(string)),
			Label:         *conv.PString(d.Get("label").(string)),
			Type:          *conv.PString(d.Get("data_type").(string)),
			ValueIsHidden: conv.PBool(d.Get("hidden").(bool)),
		},
		"id",
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if userAttributes.Id == nil {
		return diag.Errorf("user attribute %s has missing id", userAttributes.Name)
	}
	d.SetId(*userAttributes.Id)

	return resourceUserAttributeRead(ctx, d, c)
}

func resourceUserAttributeRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttributes, err := api.UserAttribute(d.Id(),
		"id,name,label,data_type,hidden,default_value,user_can_view,user_can_edit,domain_whitelist",
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("id", userAttributes.Id),
		d.Set("name", userAttributes.Name),
		d.Set("label", userAttributes.Label),
		d.Set("data_type", userAttributes.Type),
		d.Set("hidden", userAttributes.ValueIsHidden),
		d.Set("default_value", userAttributes.DefaultValue),
		d.Set("user_can_view", userAttributes.UserCanView),
		d.Set("user_can_edit", userAttributes.UserCanEdit),
		d.Set("domain_whitelist", userAttributes.HiddenValueDomainWhitelist),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserAttributeUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	return diag.Errorf("not yet implemented")
}

func resourceUserAttributeDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	return diag.Errorf("not yet implemented")
}
