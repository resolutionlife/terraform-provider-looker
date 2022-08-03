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

func resourceUserRoleCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// get user_id from resource data
	userID := d.Get("user_id").(string)
	diff, roleIDs, err := getDiff(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	_, setErr := api.SetUserRoles(userID, append(diff, roleIDs...), "", nil)
	if err != nil {
		return diag.FromErr(errors.Errorf("create set err: %s, userID: %s", setErr.Error(), userID))
	}

	// the state has the role_ids provisioned in tf already, the id is a concat of the set role_ids and user_id
	d.SetId(strings.Join(append(roleIDs, userID), "_"))

	// return nil
	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)
	roles, rolesErr := api.UserRoles(sdk.RequestUserRoles{UserId: userID}, nil)
	if rolesErr != nil {
		// TODO: Account for the case when the user is not found
		return diag.Errorf("err: %s, userID: %s", rolesErr.Error(), userID)
	}
	roleIDs, strErr := lookerRolesToSliceString(roles)
	if strErr != nil {
		return diag.FromErr(strErr)
	}

	diff, roleIDs, diffErr := getDiff(d, c)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	toSet := delete(roleIDs, diff)

	result := multierror.Append(
		d.Set("user_id", userID),
		d.Set("role_ids", toSet),
	)

	return diag.FromErr(result.ErrorOrNil())

	return nil
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

func resourceUserRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// know current config ([18]), know what was previously configured ([18, 19]), know what exists ([3, 19, 18])

	// get user_id from resource data
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
	roles, rolesErr := api.UserRoles(sdk.RequestUserRoles{
		UserId: userID,
	}, nil)
	if rolesErr != nil {
		// TODO: Account for the case when the user is not found
		return diag.Errorf("getdiff err: %s, userID: %s", rolesErr.Error(), userID)

	}

	// diff between what was set and what is in looker
	diff, diffErr := sliceRolesDiff(oldIDs, roles)
	if diffErr != nil {
		return diag.FromErr(diffErr)
	}

	// diff or old  [18,19] and what's in looke [18,19,3] =3
	// set append(3, new)

	// return diag.Errorf("old %+v, new %v, diff:%v, append: %v", oldIDs, newIDs, diff, append(diff, newIDs...))

	_, setErr := api.SetUserRoles(userID, append(diff, newIDs...), "", nil)
	if setErr != nil {
		return diag.Errorf(setErr.Error())
	}

	return resourceUserRoleRead(ctx, d, c)
}

func getRolesFromResource(d *schema.ResourceData) ([]string, diag.Diagnostics) {
	// get role_ids from resource data
	roleIDsSet, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return nil, diag.Errorf("an error occured asserting the role_id type to *Set")
	}
	roleIDs := make([]string, roleIDsSet.Len())
	for i, r := range roleIDsSet.List() {
		roleID, ok := r.(string)
		if !ok {
			return nil, diag.Errorf("attribute role_ids contains a non-string value")
		}
		roleIDs[i] = roleID
	}
	return roleIDs, nil

}

func resourceUserRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	diff, _, err := getDiff(d, c)
	if err != nil {
		return diag.FromErr(err)
	}

	// setting an empty slice of strings acts to delete the existing roles allocated to the user
	_, setErr := api.SetUserRoles(d.Get("user_id").(string), diff, "", nil)

	// TODO: Account for the case when the user
	if err != nil {
		return diag.Errorf("err: %s, userID: %s", setErr.Error(), d.Get("user_id").(string))
	}
	return nil
}

// setUserRole sets a list of roles to a user and returns the user ID and a slice of role IDs
func getDiff(d *schema.ResourceData, c interface{}) ([]string, []string, error) {
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

func sliceDiff(s, t []string) (diff []string, err error) {
	for _, sv := range s {
		if !contains(t, sv) {
			diff = append(diff, sv)
		}
	}
	return
}

func sliceRolesDiff(s []string, roles []sdk.Role) (diff []string, err error) {
	for _, role := range roles {
		if role.Id == nil {
			// TODO Make better
			return nil, errors.Errorf("the user has a role with a missing id")
		}
		if !contains(s, *role.Id) {
			diff = append(diff, *role.Id)
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
