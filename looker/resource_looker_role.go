package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"

	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manages roles of a Looker instance",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the role",
			},
			"permission_set_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The permission set id for the role. A permission set is a collection of permissions which defines what a user may do",
			},
			"model_set_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The model set id for the role. A model set is a collection of models that define what models a role can access",
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	role, roleErr := api.CreateRole(
		sdk.WriteRole{
			Name:            conv.PString(d.Get("name").(string)),
			PermissionSetId: conv.PString(d.Get("permission_set_id").(string)),
			ModelSetId:      conv.PString(d.Get("model_set_id").(string)),
		}, nil,
	)
	if roleErr != nil {
		return diag.FromErr(roleErr)
	}

	if role.Id == nil {
		return diag.Errorf("role id is nil")
	}
	d.SetId(*role.Id)

	return resourceRoleRead(ctx, d, c)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	role, roleErr := api.Role(d.Id(), nil)
	if roleErr != nil {
		// TODO: handle case when role is not found
		return diag.FromErr(roleErr)
	}

	result := multierror.Append(
		d.Set("name", role.Name),
		d.Set("permission_set_id", role.PermissionSetId),
		d.Set("model_set_id", role.ModelSetId),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	_, updateErr := api.UpdateRole(d.Id(),
		sdk.WriteRole{
			Name:            conv.PString(d.Get("name").(string)),
			PermissionSetId: conv.PString(d.Get("permission_set_id").(string)),
			ModelSetId:      conv.PString(d.Get("model_set_id").(string)),
		}, nil,
	)
	if updateErr != nil {
		diag.FromErr(updateErr)
	}

	return resourceRoleRead(ctx, d, c)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	_, delErr := api.DeleteRole(d.Id(), nil)
	if delErr != nil {
		// TODO: handle case where role is not found
		return diag.FromErr(delErr)
	}

	return nil
}
