package looker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"

	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func resourceGroupUser() *schema.Resource {
	return &schema.Resource{
		Description: "This resource adds a Looker user to a user group.",

		CreateContext: resourceGroupUserCreate,
		ReadContext:   resourceGroupUserRead,
		UpdateContext: resourceGroupUserUpdate,
		DeleteContext: resourceGroupUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupUserImport,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user group",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the looker_group_user binding. The id is of the form <user_id>_<group_id>",
			},
		},
	}
}

func resourceGroupUserCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupID := d.Get("group_id").(string)
	userID := d.Get("user_id").(string)

	_, usrErr := api.AddGroupUser(
		groupID,
		sdk.GroupIdForGroupUserInclusion{
			UserId: conv.PString(userID),
		},
		nil,
	)
	if usrErr != nil {
		return diag.FromErr(usrErr)
	}

	d.SetId(fmt.Sprintf("%s_%s", userID, groupID))

	return resourceGroupUserRead(ctx, d, c)
}

func resourceGroupUserRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userIDs, usersErr := api.SearchUsers(sdk.RequestSearchUsers{
		GroupId: conv.PString(d.Get("group_id").(string)),
		Id:      conv.PString(d.Get("user_id").(string)),
	}, nil)
	if usersErr != nil {
		return diag.FromErr(usersErr)
	}

	// userIDs will be empty if user does not belong to group
	usrIds := make([]string, len(userIDs))
	for _, usr := range userIDs {
		if usr.Id == nil {
			return diag.Errorf("the group has a user with a missing id")
		}
		usrIds = append(usrIds, *usr.Id)
	}

	if !slice.Contains(usrIds, d.Get("user_id").(string)) {
		d.SetId("")
	}

	return nil
}

func resourceGroupUserUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	oldUsr, newUsr := d.GetChange("user_id")
	oldGr, newGr := d.GetChange("group_id")

	// delete old user from old group
	delErr := api.DeleteGroupUser(oldGr.(string), oldUsr.(string), nil)
	// TODO: amend this when the looker SDK has merged this PR https://github.com/looker-open-source/sdk-codegen/pull/1074
	if delErr != nil && !errors.Is(delErr, io.EOF) {
		return diag.FromErr(delErr)
	}

	// add new user to new group
	_, addErr := api.AddGroupUser(newGr.(string),
		sdk.GroupIdForGroupUserInclusion{
			UserId: conv.PString(newUsr.(string)),
		}, nil,
	)
	if addErr != nil {
		return diag.FromErr(addErr)
	}

	d.SetId(fmt.Sprintf("%s_%s", newUsr, newGr))

	return resourceGroupUserRead(ctx, d, c)
}

func resourceGroupUserDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	delErr := api.DeleteGroupUser(d.Get("group_id").(string), d.Get("user_id").(string), nil)
	// the sdk attempts to decode the api response body into a nil struct - this is not an error from looker and can be ignored
	// TODO: amend this when the looker SDK has merged this PR https://github.com/looker-open-source/sdk-codegen/pull/1074
	if !errors.Is(delErr, io.EOF) {
		return diag.FromErr(delErr)
	}

	return nil
}

func resourceGroupUserImport(ctx context.Context, d *schema.ResourceData, c interface{}) ([]*schema.ResourceData, error) {
	// id is <user_id>_<group_id>
	s := strings.Split(d.Id(), "_")
	if len(s) < 2 {
		diag.Errorf("invalid id, should be of the form <user_id>_<group_id>")
	}

	resErr := multierror.Append(
		d.Set("user_id", s[0]),
		d.Set("group_id", s[1]),
	).ErrorOrNil()
	if resErr != nil {
		return nil, resErr
	}

	return []*schema.ResourceData{d}, nil
}
