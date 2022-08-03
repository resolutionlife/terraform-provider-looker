package looker

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
)

func resourceUserRole() *schema.Resource {
	return &schema.Resource{
		Description: "Allocates roles to a user of a Looker instance",

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

	// get user_id from resource data
	userID := d.Get("user_id").(string)
	diff, roleIDs, err := diffOfStateAndLooker(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	_, setErr := api.SetUserRoles(userID, append(diff, roleIDs...), "", nil)
	if err != nil {
		return diag.FromErr(errors.Errorf("create set err: %s, userID: %s", setErr.Error(), userID))
	}

	// the state has the role_ids provisioned in tf already, the id is a concat of the set role_ids and user_id
	d.SetId(strings.Join(append(roleIDs, userID), "_"))

	return resourceUserRoleRead(ctx, d, c)
}

// resourceUserRoleRead reads what has been set in looker, and sets just what is provisioned in terraform to the state
func resourceUserRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	// diff of what is in current state and in looker
	diff, lookerRoleIDs, diffErr := diffOfStateAndLooker(d, c)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	result := multierror.Append(
		d.Set("user_id", d.Get("user_id").(string)),
		d.Set("role_ids", delete(lookerRoleIDs, diff)), // delete diff from lookerRoleIDs to only set IDs that have been provisioned in the provider
	)
	return diag.FromErr(result.ErrorOrNil())
}

// resourceUserRoleUpdate gets looks at the diff between the old and new state, and appends these changes to the existing roles in looker
func resourceUserRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)

	o, n := d.GetChange("role_ids")
	// IDs set in old state
	oldIDs, oErr := setToSliceString(o)
	if oErr != nil {
		return diag.FromErr(oErr)
	}
	// IDs set in new state
	newIDs, nErr := setToSliceString(n)
	if nErr != nil {
		return diag.FromErr(nErr)
	}

	// get role_ids that exist already
	ur, urErr := api.UserRoles(sdk.RequestUserRoles{UserId: userID}, nil)
	if urErr != nil {
		// TODO: Account for the case when the user is not found
		return diag.FromErr(urErr)

	}
	lookerRoles, rolesErr := lookerRolesToSliceString(ur)
	if rolesErr != nil {
		diag.FromErr(rolesErr)
	}

	// diff between what was has changed in the state and what is in looker
	diff, diffErr := sliceDiff(oldIDs, lookerRoles)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	/*
		eg.
			new =    [developer],
			old =    [viewer, developer],
			looker = [viewer, developer, user],
			diff =   [user]
			toSet =  [user, developer]
	*/

	_, setErr := api.SetUserRoles(userID, append(diff, newIDs...), "", nil)
	if setErr != nil {
		return diag.Errorf(setErr.Error())
	}

	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)

	diff, _, diffErr := diffOfStateAndLooker(d, c)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	// TODO: Account for the case when the user
	_, setErr := api.SetUserRoles(userID, diff, "", nil)
	if setErr != nil {
		return diag.Errorf("err: %s, userID: %s", setErr.Error(), d.Get("user_id").(string))
	}

	return nil
}

// diffOfStateAndLooker returns the diff between the looker roles and what is set in the state.
// This function returns the diff and the roles set in the state.
func diffOfStateAndLooker(d *schema.ResourceData, c interface{}) ([]string, []string, error) {
	api := c.(*sdk.LookerSDK)

	// get user_id from resource data
	userID := d.Get("user_id").(string)

	// get role_ids from resource data
	rscRoleIDs, setErr := setToSliceString(d.Get("role_ids"))
	if setErr != nil {
		return nil, nil, setErr
	}

	// get role_ids that exist already
	ur, urErr := api.UserRoles(sdk.RequestUserRoles{UserId: userID}, nil)
	if urErr != nil {
		// TODO: Account for the case when the user is not found
		return nil, nil, errors.Errorf("getdiff err: %s, userID: %s", urErr.Error(), userID)
	}
	lookerRoles, lErr := lookerRolesToSliceString(ur)
	if lErr != nil {
		return nil, nil, lErr
	}

	diff, diffErr := sliceDiff(rscRoleIDs, lookerRoles)
	if diffErr != nil {
		return nil, nil, diffErr
	}

	return diff, rscRoleIDs, nil
}

func lookerRolesToSliceString(roles []sdk.Role) ([]string, error) {
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		if role.Id == nil {
			return nil, errors.Errorf("the user has a role with a missing id")
		}
		roleIDs[i] = *role.Id
	}

	return roleIDs, nil
}

func setToSliceString(i interface{}) ([]string, error) {
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

func sliceDiff(s, t []string) (diff []string, err error) {
	for _, sv := range s {
		if !contains(t, sv) {
			diff = append(diff, sv)
		}
	}
	return
}

func contains(s []string, v string) bool {
	for i := range s {
		if s[i] == v {
			return true
		}
	}
	return false
}

func delete(s []string, toDelete []string) (str []string) {
	for i := range s {
		if !contains(toDelete, s[i]) {
			str = append(str, s[i])
		}
	}
	return
}
