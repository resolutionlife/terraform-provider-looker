package looker

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"

	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func resourceUserRole() *schema.Resource {
	return &schema.Resource{
		Description: "This resource binds a set of roles to a looker user. This is an additive and non-authorative resource that grants roles in addition to current roles configured in Looker.",

		CreateContext: resourceUserRoleCreate,
		ReadContext:   resourceUserRoleRead,
		UpdateContext: resourceUserRoleUpdate,
		DeleteContext: resourceUserRoleDelete,
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

// resourceUserRoleCreate reads what exists in looker and appends the new roles to the existing roles
func resourceUserRoleCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// get diff between roles in the resource data and in looker
	diff, err := userRolesDiff(api, d)
	if err != nil {
		return diag.FromErr(err)
	}

	userID := d.Get("user_id").(string)
	rscRoleIDs, rolesErr := getRolesByUser(api, userID)
	if rolesErr != nil {
		diag.FromErr(rolesErr)
	}

	_, setErr := api.SetUserRoles(userID, append(diff, rscRoleIDs...), "", nil)
	if err != nil {
		return diag.FromErr(errors.Errorf("create set err: %s, userID: %s", setErr.Error(), userID))
	}

	// the state has the role_ids provisioned in tf already, the id is a concat of the set role_ids and user_id
	d.SetId(strings.Join(append(rscRoleIDs, userID), "_"))

	return resourceUserRoleRead(ctx, d, c)
}

// resourceUserRoleRead reads what has been set in looker, and sets just what is provisioned in terraform to the state
func resourceUserRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	diff, diffErr := userRolesDiff(api, d)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	userID := d.Get("user_id").(string)
	rscRoleIDs, rolesErr := getRolesByUser(api, userID)
	if rolesErr != nil {
		diag.FromErr(rolesErr)
	}

	result := multierror.Append(
		d.Set("user_id", userID),
		d.Set("role_ids", slice.Delete(rscRoleIDs, diff)), // delete any role ids that are not in the terraform state
	)
	return diag.FromErr(result.ErrorOrNil())
}

// resourceUserRoleUpdate inspects the changes between the old and new state, and appends these changes to existing roles in looker for the given user_id.
func resourceUserRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	/*
		compare old and new state, and update the roles accordingly, preserving any configuration in looker only. eg.
			newIDs =    ["developer"],
			oldIDs =    ["viewer", "developer"],
			lookerRoles = ["viewer", "developer", "user"],
			diff =   ["user"]
			toSet =  ["user", ""developer""]
	*/
	o, n := d.GetChange("role_ids")
	oldIDs, oErr := schemaSetToSliceString(o)
	if oErr != nil {
		return diag.FromErr(oErr)
	}
	newIDs, nErr := schemaSetToSliceString(n)
	if nErr != nil {
		return diag.FromErr(nErr)
	}

	// get role_ids that exist already
	userID := d.Get("user_id").(string)
	lookerRoles, rolesErr := getRolesByUser(api, userID)
	if rolesErr != nil {
		diag.FromErr(rolesErr)
	}

	// diff between what was has changed in the state and what is in looker
	diff := slice.Diff(oldIDs, lookerRoles)

	_, setErr := api.SetUserRoles(userID, append(diff, newIDs...), "", nil)
	if setErr != nil {
		return diag.Errorf(setErr.Error())
	}

	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	diff, diffErr := userRolesDiff(api, d)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	userID := d.Get("user_id").(string)

	// TODO: Account for the case when the user
	_, setErr := api.SetUserRoles(userID, diff, "", nil)
	if setErr != nil {
		return diag.Errorf("err: %s, userID: %s", setErr.Error(), userID)
	}

	return nil
}

// userRolesDiff returns the diff between the looker remote roles and roles in the state.
func userRolesDiff(api *sdk.LookerSDK, d *schema.ResourceData) ([]string, error) {
	// get role ids set in the resource data
	rscRoleIDs, rscErr := schemaSetToSliceString(d.Get("role_ids"))
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

// getRolesByUser takes a client and a userID and returns a slice the roles allocated to a user with the given userID
func getRolesByUser(api *sdk.LookerSDK, userID string) ([]string, error) {
	ur, urErr := api.UserRoles(sdk.RequestUserRoles{
		UserId:                userID,
		DirectAssociationOnly: conv.PBool(true),
	}, nil)
	if urErr != nil {
		// TODO: Account for the case when the user is not found
		return nil, urErr
	}

	roleIDs := make([]string, len(ur))
	for i, role := range ur {
		if role.Id == nil {
			return nil, errors.Errorf("the user has a role with a missing id")
		}
		roleIDs[i] = *role.Id
	}

	return roleIDs, nil
}

func schemaSetToSliceString(i interface{}) ([]string, error) {
	set, ok := i.(*schema.Set)
	if !ok {
		return nil, errors.New("interface{} is not of type *schema.Set")
	}

	slice := make([]string, set.Len())
	for i, v := range set.List() {
		str, ok := v.(string)
		if !ok {
			return nil, errors.New("set contains a non-string element")
		}
		slice[i] = str
	}

	return slice, nil
}
