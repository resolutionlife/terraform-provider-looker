package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func resourceUserRole() *schema.Resource {
	return &schema.Resource{
		Description: "Allocates users to a role of a Looker instance",

		CreateContext: resourceUserRoleCreate,
		ReadContext:   resourceUserRoleRead,
		UpdateContext: resourceUserRoleUpdate,
		DeleteContext: resourceUserRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				// Description: "The permission set id for the role. A permission set is a collection of permissions which defines what a user may do",
			},
			"role_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				// Description: "The name of the role",
			},
		},
	}
}

func resourceUserRoleCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)
	roleIDsSet, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("an error occured asserting the role_id type to *Set")
	}

	roleIDs := make([]string, roleIDsSet.Len())
	for i, r := range roleIDsSet.List() {
		roleID, ok := r.(string)
		if !ok {
			return diag.Errorf("attribute role_ids contains a non-string value")
		}
		roleIDs[i] = roleID
	}

	_, setErr := api.SetUserRoles(userID, roleIDs, "id", nil)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	d.SetId(userID)

	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	roles, rolesErr := api.UserRoles(sdk.RequestUserRoles{
		UserId: d.Id(),
	}, nil)
	if rolesErr != nil {
		// TODO: Account for the case when the user is not found
		return diag.FromErr(rolesErr)
	}

	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		if role.Id == nil {
			return diag.Errorf("the user with id %s has a role with a missing id", d.Id())
		}
		roleIDs[i] = *role.Id
	}
	result := multierror.Append(
		d.Set("user_id", d.Id()),
		d.Set("role_ids", roleIDs),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserRoleUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)
	roleIDsSet, ok := d.Get("role_ids").(*schema.Set)
	if !ok {
		return diag.Errorf("an error occured asserting the role_id type to *Set")
	}

	roleIDs := make([]string, roleIDsSet.Len())
	for i, r := range roleIDsSet.List() {
		roleID, ok := r.(string)
		if !ok {
			return diag.Errorf("attribute role_ids contains a non-string value")
		}
		roleIDs[i] = roleID
	}

	_, setErr := api.SetUserRoles(userID, roleIDs, "id", nil)
	if setErr != nil {
		return diag.FromErr(setErr)
	}

	d.SetId(userID)

	return resourceUserRoleRead(ctx, d, c)
}

func resourceUserRoleDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// set no user roles deletes the user role binding
	_, setErr := api.SetUserRoles(
		d.Get("user_id").(string),
		[]string{},
		"", nil,
	)

	// TODO: Account for the case when the user is not found
	return diag.FromErr(setErr)
}
