package looker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"

	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func resourceRoleGroups() *schema.Resource {
	return &schema.Resource{
		Description: "This resource binds a set of groups to a looker role. This is an additive and non-authorative resource that grants groups to a role in addition to current groups configured in Looker.",

		CreateContext: resourceRoleGroupsCreate,
		ReadContext:   resourceRoleGroupsRead,
		UpdateContext: resourceRoleGroupsUpdate,
		DeleteContext: resourceRoleGroupsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRoleGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the role",
				ForceNew:    true,
			},
			"group_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "An unordered list of group ids to be granted the role",
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
	if errors.Is(setErr, sdk.ErrNotFound) {
		return diag.Errorf("role with id %s cannot be found", roleID)
	}
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	d.SetId(fmt.Sprintf("%s_%s", roleID, strings.Join(groupIDs, "_")))

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

	// this method only handles group_id changes as resource is replaced if role_id changes
	// get old and new groups set on this resource
	o, n := d.GetChange("group_ids")
	old, ok := o.(*schema.Set)
	if !ok {
		return diag.Errorf("old role_id is not of type *schema.Set")
	}
	oldIDs, oErr := conv.SchemaSetToSliceString(old)
	if oErr != nil {
		return diag.FromErr(oErr)
	}
	new, ok := n.(*schema.Set)
	if !ok {
		return diag.Errorf("new role_id is not of type *schema.Set")
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
	diff := slice.LeftDiff(oldIDs, lookerGroupIDs)
	_, setErr := api.SetRoleGroups(roleID, append(diff, newIDs...), nil)
	if errors.Is(setErr, sdk.ErrNotFound) {
		return diag.Errorf("role with id %s cannot be found", roleID)
	}
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	d.SetId(fmt.Sprintf("%s_%s", roleID, strings.Join(newIDs, "_")))

	return resourceRoleGroupsRead(ctx, d, c)
}

func resourceRoleGroupsDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	gIDs, ok := d.Get("group_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("attribute role_id is not of type *schema.Set")
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
	if !errors.Is(setErr, sdk.ErrNotFound) {
		return diag.FromErr(setErr)
	}

	return nil
}

func resourceRoleGroupImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// id is delimited using `_`, eg. <role_id>_<group_ids>
	s := strings.Split(d.Id(), "_")
	if len(s) < 2 {
		diag.Errorf("invalid id, should be of the form <role_id>_<group_ids>")
	}

	resErr := multierror.Append(
		d.Set("role_id", s[0]),
		d.Set("group_ids", s[1:]),
	).ErrorOrNil()
	if resErr != nil {
		return nil, resErr
	}

	return []*schema.ResourceData{d}, nil
}

func getGroupsOnRole(api *sdk.LookerSDK, roleID string) ([]string, error) {
	g, gErr := api.RoleGroups(roleID, "id", nil)
	if errors.Is(gErr, sdk.ErrNotFound) {
		return nil, fmt.Errorf("role with id %s cannot be found", roleID)
	}
	if gErr != nil {
		return nil, gErr
	}

	// g will be empty if there are no groups with the given role
	lookerGroupIDs := make([]string, len(g))
	for i, group := range g {
		if group.Id == nil {
			return nil, fmt.Errorf("the user has a group with a missing id")
		}
		lookerGroupIDs[i] = *group.Id
	}

	return lookerGroupIDs, nil
}
