package looker

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
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
	userID, roleIDs, errDiag := setUserRole(d, c)
	if errDiag != nil {
		return errDiag
	}

	d.SetId(strings.Join(append(roleIDs, userID), "_"))

	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)
	roles, rolesErr := api.UserRoles(sdk.RequestUserRoles{
		UserId: userID,
	}, nil)
	if rolesErr != nil {
		// TODO: Account for the case when the user is not found
		return diag.FromErr(rolesErr)
	}

	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		if role.Id == nil {
			return diag.Errorf("the user with id %s has a role with a missing id", userID)
		}
		roleIDs[i] = *role.Id
	}
	result := multierror.Append(
		// d.Set("user_id", userID),
		d.Set("role_ids", roleIDs),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	_, _, errDiag := setUserRole(d, c)
	if errDiag != nil {
		return errDiag
	}

	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// setting an empty slice of strings acts to delete the existing roles allocated to the user
	_, setErr := api.SetUserRoles(d.Get("user_id").(string), []string{}, "", nil)

	// TODO: Account for the case when the user is not found
	return diag.FromErr(setErr)
}

// setUserRole sets a list of roles to a user and returns the user ID and a slice of role IDs
func setUserRole(d *schema.ResourceData, c interface{}) (string, []string, diag.Diagnostics) {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)
	roleIDsSet, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return "", nil, diag.Errorf("an error occured asserting the role_id type to *Set")
	}

	roleIDs := make([]string, roleIDsSet.Len())
	for i, r := range roleIDsSet.List() {
		roleID, ok := r.(string)
		if !ok {
			return "", nil, diag.Errorf("attribute role_ids contains a non-string value")
		}
		roleIDs[i] = roleID
	}

	_, setErr := api.SetUserRoles(userID, roleIDs, "", nil)
	if setErr != nil {
		return "", nil, diag.FromErr(setErr)
	}

	return userID, roleIDs, nil
}
