package looker

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"

	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func resourceUserRoles() *schema.Resource {
	return &schema.Resource{
		Description: "This resource binds a set of roles to a Looker user. This is an additive and non-authorative resource that grants roles in addition to current roles configured in Looker.",

		CreateContext: resourceUserRolesCreate,
		ReadContext:   resourceUserRolesRead,
		UpdateContext: resourceUserRolesUpdate,
		DeleteContext: resourceUserRolesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user",
			},
			"role_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "A slice of role_ids which will be assigned to the user",
			},
		},
	}
}

// resourceUserRolesCreate reads what exists in looker and appends the new roles to the existing roles
func resourceUserRolesCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// get diff between roles in the resource data and in looker
	diff, err := userRolesDiff(api, d)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	rscRoleIDs, rolesErr := getRolesByUser(api, userID)
	if rolesErr != nil {
		return diag.FromErr(rolesErr)
	}

	_, setErr := api.SetUserRoles(userID, append(diff, rscRoleIDs...), "", nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("create set err: %s, userID: %s", setErr.Error(), userID))
	}

	d.SetId(userID)

	return resourceUserRolesRead(ctx, d, c)
}

// resourceUserRolesRead reads what has been set in looker, and sets just what is provisioned in terraform to the state
func resourceUserRolesRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	diff, diffErr := userRolesDiff(api, d)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	rscRoleIDs, rolesErr := getRolesByUser(api, d.Id())
	if rolesErr != nil {
		return diag.FromErr(rolesErr)
	}

	result := multierror.Append(
		d.Set("user_id", d.Id()),
		d.Set("role_ids", slice.Delete(rscRoleIDs, diff)), // delete any role ids that are not in the terraform state
	)
	return diag.FromErr(result.ErrorOrNil())
}

// resourceUserRolesUpdate inspects the changes between the old and new state, and appends these changes to existing roles in looker for the given user_id.
func resourceUserRolesUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	/*
		compare old and new state, and update the roles accordingly, preserving any configuration in looker only. eg.
			newIDs =    ["developer"],
			oldIDs =    ["viewer", "developer"],
			lookerRoles = ["viewer", "developer", "user"],
			diff =   ["user"]
			toSet =  ["user", "developer"]
	*/

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

	// get role_ids that exist already
	userID := d.Get("user_id").(string)
	lookerRoles, rolesErr := getRolesByUser(api, userID)
	if rolesErr != nil {
		return diag.FromErr(rolesErr)
	}

	// diff between what was has changed in the state and what is in looker
	diff := slice.Diff(oldIDs, lookerRoles)

	_, setErr := api.SetUserRoles(userID, append(diff, newIDs...), "", nil)
	if setErr != nil {
		return diag.Errorf(setErr.Error())
	}

	return resourceUserRolesRead(ctx, d, c)
}

func resourceUserRolesDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	diff, diffErr := userRolesDiff(api, d)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	_, setErr := api.SetUserRoles(d.Id(), diff, "", nil)
	if setErr != nil {
		return diag.Errorf("err: %s, userID: %s", setErr.Error(), d.Id())
	}

	return nil
}

// userRolesDiff returns the diff between the looker remote roles and roles in the state.
func userRolesDiff(api *sdk.LookerSDK, d *schema.ResourceData) ([]string, error) {
	// get role ids set in the resource data
	rIDs, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return nil, errors.New("role_ids is not of type *schema.Set")
	}
	rscRoleIDs, rscErr := conv.SchemaSetToSliceString(rIDs)
	if rscErr != nil {
		return nil, rscErr
	}

	// get role ids that exist already for a given user
	userID := d.Get("user_id").(string)
	lookerRoleIDs, lErr := getRolesByUser(api, userID)
	if lErr != nil {
		return nil, lErr
	}

	return slice.Diff(rscRoleIDs, lookerRoleIDs), nil
}

// getRolesByUser takes a client and a userID and returns a slice the roles allocated to a user with the given userID.
func getRolesByUser(api *sdk.LookerSDK, userID string) ([]string, error) {
	ur, urErr := api.UserRoles(sdk.RequestUserRoles{
		UserId:                userID,
		DirectAssociationOnly: conv.PBool(true),
	}, nil)
	if urErr != nil {
		return nil, urErr
	}

	// if no role is set on the user, ur will be an empty slice
	roleIDs := make([]string, len(ur))
	for i, role := range ur {
		if role.Id == nil {
			return nil, errors.New("the user has a role with a missing id")
		}
		roleIDs[i] = *role.Id
	}

	return roleIDs, nil
}
