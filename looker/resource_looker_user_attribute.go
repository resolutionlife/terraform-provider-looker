package looker

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
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
				Type:             schema.TypeString,
				Required:         true,
				Description:      "",
				ValidateDiagFunc: validationFunc([]string{"string", "number", "datetime", "yesno", "zipcode"}),
			},
			"hidden": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "If set the value will be treated like a password, and once set, no one will be able to decrypt and view it",
			},
			"user_access": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "If 'None', non-admins will not be able to see the value of this attribute for themselves. If 'View' is required to use this attribute in query filters. If 'Edit', the user will be able to set their own value of this attribute, so the user attribute will not be able to be used as an access filter.",
				ValidateDiagFunc: validationFunc([]string{"None", "View", "Edit"}),
			},
			"default_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value when no other value is set for the user or for one of the user's groups",
			},
			"domain_whitelist": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "A list of urls that will be allowed as a destination for this user attribute, optionally using a wildcard '*'. You must set this when changing a user attribute to 'hidden'. Once set values can only be changed to be more restrictive. (I.e. removing elements from the list or changing an entry like 'my_domain/*' to 'my_domain/route/*')",
			},
		},
	}
}

func resourceUserAttributeCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttrs, err := buildUserAttributeInput(d)
	if err != nil {
		return diag.FromErr(err)
	}

	userAttributes, err := api.CreateUserAttribute(*userAttrs, "id", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	if userAttributes.Id == nil {
		return diag.Errorf("user attribute has missing id")
	}
	d.SetId(*userAttributes.Id)

	return resourceUserAttributeRead(ctx, d, c)
}

func resourceUserAttributeRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttributes, err := api.UserAttribute(d.Id(), "", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var defaultValue *string
	// if hidden, default value cannot be retrieved from API
	if *userAttributes.ValueIsHidden {
		// use has change since API cannot give us default value
		if d.HasChange("default_value") {
			defaultValue = conv.PString("DEFAULT_IS_SET")
		} else {
			defaultValue = conv.PString(d.Get("default_value").(string))
		}
	} else {
		defaultValue = userAttributes.DefaultValue
	}

	var userAccess string
	if *userAttributes.UserCanView {
		userAccess = "View"
	} else if *userAttributes.UserCanEdit {
		userAccess = "Edit"
	} else {
		userAccess = "None"
	}

	var domainsWhitelistSlice []string
	if userAttributes.HiddenValueDomainWhitelist != nil {
		domainsWhitelistSlice = strings.Split(*userAttributes.HiddenValueDomainWhitelist, ",")
	}

	result := multierror.Append(
		d.Set("id", userAttributes.Id),
		d.Set("name", userAttributes.Name),
		d.Set("label", userAttributes.Label),
		d.Set("data_type", userAttributes.Type),
		d.Set("hidden", userAttributes.ValueIsHidden),
		d.Set("default_value", defaultValue),
		d.Set("user_access", conv.PString(userAccess)),
		d.Set("domain_whitelist", conv.PSlices(domainsWhitelistSlice)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserAttributeUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttrs, err := buildUserAttributeInput(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.UpdateUserAttribute(d.Id(), *userAttrs, "id", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePermissionSetRead(ctx, d, c)
}

func resourceUserAttributeDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	return diag.Errorf("not yet implemented")
}

func buildUserAttributeInput(d *schema.ResourceData) (*sdk.WriteUserAttribute, error) {
	userAttr := sdk.WriteUserAttribute{
		Name:          *conv.PString(d.Get("name").(string)),
		Label:         *conv.PString(d.Get("label").(string)),
		Type:          *conv.PString(d.Get("data_type").(string)),
		ValueIsHidden: conv.PBool(d.Get("hidden").(bool)),
		DefaultValue:  conv.PString(d.Get("default_value").(string)),
	}

	user_access, ok := d.Get("user_access").(string)
	if !ok {
		return nil, errors.New("user_access is not a string")
	}

	switch user_access {
	case "View":
		userAttr.UserCanView = conv.PBool(true)
	case "Edit":
		userAttr.UserCanEdit = conv.PBool(true)
	case "None":
		userAttr.UserCanView = conv.PBool(false)
		userAttr.UserCanEdit = conv.PBool(false)
	}

	domainsWhitelist, ok := d.Get("domain_whitelist").(*schema.Set)
	if !ok {
		return nil, errors.New("domain_whitelist is not a set")
	}

	domainsWhitelistSlice, err := conv.SchemaSetToSliceString(domainsWhitelist)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert domain_whitelist to string slice")
	}

	domainWhitelistString := strings.Join(domainsWhitelistSlice, ",")
	if domainWhitelistString != "" {
		userAttr.HiddenValueDomainWhitelist = conv.PString(domainWhitelistString)
	}

	return &userAttr, nil
}

//Useful tools, maybe could be reused by other resources, need to find good home for these
func validationFunc[T comparable](validOptions []T) func(i interface{}, p cty.Path) diag.Diagnostics {
	return func(i interface{}, p cty.Path) diag.Diagnostics {
		value := i.(T)

		var diags diag.Diagnostics
		if !contains(validOptions, value) {
			diag := diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "invalid option",
				Detail:   fmt.Sprintf("%v is not included in list of valid options: %v", value, validOptions),
			}
			diags = append(diags, diag)
		}
		return diags
	}
}

func contains[T comparable](s []T, v T) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}