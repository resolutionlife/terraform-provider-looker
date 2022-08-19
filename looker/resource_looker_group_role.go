package looker

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceGroupRole() *schema.Resource {
	return &schema.Resource{
		Description: "",

		CreateContext: resourceGroupRoleCreate,
		ReadContext:   resourceGroupRoleRead,
		UpdateContext: resourceGroupRoleUpdate,
		DeleteContext: resourceGroupRoleDelete,
		Importer: &schema.ResourceImporter{
			// todo
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"role_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "",
			},
		},
	}
}

func resourceGroupRoleCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupID := d.Get("group_id").(string)
	rIDs, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("attribute role_ids is not of type *schema.Set")
	}
	roleIDs, rErr := conv.SchemaSetToSliceString(rIDs)
	if rErr != nil {
		return diag.FromErr(rErr)
	}

	var result *multierror.Error
	for _, roleID := range roleIDs {
		_, setErr := api.SetRoleGroups(roleID, []string{groupID}, nil)
		if setErr != nil {
			multierror.Append(
				result, setErr,
			)
		}
	}
	if resErr := result.ErrorOrNil(); resErr != nil {
		diag.FromErr(resErr)
	}

	d.SetId(strings.Join(append(roleIDs, groupID), "_"))

	return resourceGroupRoleRead(ctx, d, c)
}
func resourceGroupRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupID := d.Get("group_id").(string)

	res, rErr := api.SearchGroupsWithRoles(sdk.RequestSearchGroups{
		Id: conv.PString(groupID),
	}, nil)
	if rErr != nil {
		return diag.FromErr(rErr)
	}

	if len(res) != 1 {
		return diag.Errorf("group with group id %s not found", groupID)
	}

	if res[0].Roles == nil {
		return diag.Errorf("no roles found on group with group id %s", groupID)
	}

	roleIDs := make([]string, len(*res[0].Roles))
	for i, role := range *res[0].Roles {
		if role.Id == nil {
			return diag.Errorf("no role id found")
		}
		roleIDs[i] = *role.Id
	}

	return diag.FromErr(d.Set("role_ids", roleIDs))
}

func resourceGroupRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupID := d.Get("group_id").(string)
	o, n := d.GetChange("role_ids")
	old, ok := o.(*schema.Set)
	if !ok {
		return diag.Errorf("old role_ids is not of type *schema.Set")
	}
	oldIDs, oErr := conv.SchemaSetToSliceString(old)
	if oErr != nil {
		return diag.FromErr(oErr)
	}

	new, ok := n.(*schema.Set)
	if !ok {
		return diag.Errorf("new role_ids is not of type *schema.Set")
	}
	newIDs, nErr := conv.SchemaSetToSliceString(new)
	if nErr != nil {
		return diag.FromErr(nErr)
	}

	// return diag.Errorf("old: %v, new: %v", oldIDs, newIDs)

	var result *multierror.Error

	// delete old roleIDs
	for _, roleID := range oldIDs {
		_, setErr := api.SetRoleGroups(roleID, []string{}, nil)
		if setErr != nil {
			multierror.Append(
				result, setErr,
			)
		}
	}

	// add new roleIDs
	for _, roleID := range newIDs {
		_, setErr := api.SetRoleGroups(roleID, []string{groupID}, nil)
		if setErr != nil {
			multierror.Append(
				result, setErr,
			)
		}
	}
	if resErr := result.ErrorOrNil(); resErr != nil {
		diag.FromErr(resErr)
	}

	d.SetId(strings.Join(append(newIDs, groupID), "_"))

	return resourceGroupRoleRead(ctx, d, c)
}

func resourceGroupRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	rIDs, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("attribute role_ids is not of type *schema.Set")
	}
	roleIDs, rErr := conv.SchemaSetToSliceString(rIDs)
	if rErr != nil {
		return diag.FromErr(rErr)
	}

	var result *multierror.Error
	for _, roleID := range roleIDs {
		_, setErr := api.SetRoleGroups(roleID, []string{}, nil)
		if setErr != nil {
			multierror.Append(
				result, setErr,
			)
		}
	}

	return nil
}

func setRolesOnGroup(api *sdk.LookerSDK, roleIDs []string, groupIDs []string) error {
	var result *multierror.Error

	for _, roleID := range roleIDs {
		_, setErr := api.SetRoleGroups(roleID, groupIDs, nil)
		if setErr != nil {
			multierror.Append(
				result, setErr,
			)
		}
	}

	return result.ErrorOrNil()
}
