package looker

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func resourceRoleGroups() *schema.Resource {
	return &schema.Resource{
		Description: "",

		CreateContext: resourceRoleGroupsCreate,
		ReadContext:   resourceRoleGroupsRead,
		UpdateContext: resourceRoleGroupsUpdate,
		DeleteContext: resourceRoleGroupsDelete,
		Importer: &schema.ResourceImporter{
			// todo
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"role_ids": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"group_ids": {
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

func resourceRoleGroupsCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	roleID := d.Get("role_id").(string)

	// read groups that are set on the role
	lookerGroupIDs, groupsErr := getGroupsOnRole(api, roleID)
	if groupsErr != nil {
		return diag.FromErr(groupsErr)
	}

	// read groups to be set by this resource
	gIDs, ok := d.Get("group_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("attribute group_ids is not of type *schema.Set")
	}
	groupIDs, gErr := conv.SchemaSetToSliceString(gIDs)
	if gErr != nil {
		return diag.FromErr(gErr)
	}

	// set all groups on the role
	_, setErr := api.SetRoleGroups(roleID,
		append(lookerGroupIDs, slice.Diff(lookerGroupIDs, groupIDs)...), // append diff to the list of groups already set in looker
		nil,
	)
	if setErr != nil {
		diag.FromErr(setErr)
	}

	d.SetId(fmt.Sprintf("%s_%v", roleID, strings.Join(groupIDs, "_")))

	return resourceRoleGroupsRead(ctx, d, c)
}

func resourceRoleGroupsRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// read groups that are set on the role
	roleID := d.Get("role_id").(string)
	lookerGroupIDs, groupsErr := getGroupsOnRole(api, roleID)
	if groupsErr != nil {
		return diag.FromErr(groupsErr)
	}

	// read groups to be set by this resource
	gIDs, ok := d.Get("group_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("attribute group_ids is not of type *schema.Set")
	}
	groupIDs, gErr := conv.SchemaSetToSliceString(gIDs)
	if gErr != nil {
		return diag.FromErr(gErr)
	}

	diff := slice.Diff(lookerGroupIDs, groupIDs)

	result := multierror.Append(
		d.Set("role_id", roleID),
		d.Set("group_ids", slice.Delete(lookerGroupIDs, diff)), // delete any group ids that are not in the terraform state
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceRoleGroupsUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	o, n := d.GetChange("group_ids")

	// get old and new groups set on this resource
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

	// read groups already set on this role
	roleID := d.Get("role_id").(string)
	lookerGroupIDs, groupsErr := getGroupsOnRole(api, roleID)
	if groupsErr != nil {
		return diag.FromErr(groupsErr)
	}

	// diff between what was has changed in the state and what is in looker
	diff := slice.Diff(oldIDs, lookerGroupIDs)

	// append
	_, setErr := api.SetRoleGroups(roleID, append(diff, newIDs...), nil)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	d.SetId(strings.Join(append(newIDs, roleID), "_"))

	return resourceRoleGroupsRead(ctx, d, c)
}

func resourceRoleGroupsDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// groups
	gIDs, ok := d.Get("group_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("attribute role_ids is not of type *schema.Set")
	}
	groupIDs, rErr := conv.SchemaSetToSliceString(gIDs)
	if rErr != nil {
		return diag.FromErr(rErr)
	}

	roleID := d.Get("role_id").(string)
	lookerGroupIDs, groupsErr := getGroupsOnRole(api, roleID)
	if groupsErr != nil {
		return diag.FromErr(groupsErr)
	}

	_, setErr := api.SetRoleGroups(roleID, slice.Diff(lookerGroupIDs, groupIDs), nil)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	return nil
}

func getGroupsOnRole(api *sdk.LookerSDK, roleID string) ([]string, error) {
	g, gErr := api.RoleGroups(roleID, "id", nil)
	if gErr != nil {
		// TODO: Account for when roleID is not found
		return nil, gErr
	}

	lookerGroupIDs := make([]string, len(g))
	for i, group := range g {
		if group.Id == nil {
			return nil, errors.Errorf("the user has a group with a missing id")
		}
		lookerGroupIDs[i] = *group.Id
	}

	return lookerGroupIDs, nil
}
