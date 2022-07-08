package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

var roleResource = schema.Resource{
	Description: "Manages roles of a Looker instance",

	// TODO: Descriptions
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The email address of the user",
		},
		"permission_set_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The first name of the user",
		},
		"model_set_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The last name of the user",
		},
	},

	CreateContext: resourceRoleCreate,
	ReadContext:   resourceRoleRead,
	// UpdateContext: resourceUserUpdate,
	// DeleteContext: resourceUserDelete,
	Importer: &schema.ResourceImporter{
		StateContext: schema.ImportStatePassthroughContext,
	},
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(sdk.LookerSDK)

	role, roleErr := api.CreateRole(
		sdk.WriteRole{
			Name:            pString(d.Get("name").(string)),
			PermissionSetId: pString(d.Get("permission_set_id").(string)),
			ModelSetId:      pString(d.Get("model_set_id").(string)),
		}, nil,
	)
	if roleErr != nil {
		return diag.FromErr(roleErr)
	}
	if role.Id == nil {
		// TODO: Fix
		return diag.Errorf("role id is nil")
	}
	d.SetId(*role.Id)

	return resourceRoleRead(ctx, d, c)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(sdk.LookerSDK)

	role, roleErr := api.Role(d.Id(), nil)
	if roleErr != nil {
		// TODO: Handle case when role is not found
		return diag.FromErr(roleErr)
	}

	result := multierror.Append(
		d.Set("name", role.Name),
		d.Set("permission_set_id", role.PermissionSetId),
		d.Set("model_set_id", role.ModelSetId),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func pString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
