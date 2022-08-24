package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a set of Looker permissions",
		CreateContext: resourcePermissionSetCreate,
		ReadContext:   resourcePermissionSetRead,
		UpdateContext: resourcePermissionSetUpdate,
		DeleteContext: resourcePermissionSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of the permission set",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the permission set",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the permission set",
			},
			"permissions": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "The list of permissions in the permission set",
			},
		},
	}
}

func resourcePermissionSetCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	permissions, ok := d.Get("permissions").(*schema.Set)
	if !ok {
		return diag.Errorf("permissions is not a set")
	}

	permissionsSlice, err := conv.SchemaSetToSliceString(permissions)
	if err != nil {
		return diag.FromErr(err)
	}

	permissionSet, err := api.CreatePermissionSet(
		sdk.WritePermissionSet{
			Name:        conv.PString(d.Get("name").(string)),
			Permissions: conv.PSlices(permissionsSlice),
		},
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if permissionSet.Id == nil {
		return diag.Errorf("permission set has missing id")
	}
	d.SetId(*permissionSet.Id)

	return resourcePermissionSetRead(ctx, d, c)
}

func resourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	permissionSet, err := api.PermissionSet(d.Id(), "id,name,permissions", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*permissionSet.Id)
	result := multierror.Append(
		d.Set("name", permissionSet.Name),
		d.Set("permissions", permissionSet.Permissions),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourcePermissionSetUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	permissions, ok := d.Get("permissions").(*schema.Set)
	if !ok {
		return diag.Errorf("permissions is not a set")
	}

	permissionsSlice, err := conv.SchemaSetToSliceString(permissions)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.UpdatePermissionSet(
		d.Id(),
		sdk.WritePermissionSet{
			Name:        conv.PString(d.Get("name").(string)),
			Permissions: conv.PSlices(permissionsSlice),
		},
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePermissionSetRead(ctx, d, c)
}

func resourcePermissionSetDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	_, err := api.DeletePermissionSet(d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
