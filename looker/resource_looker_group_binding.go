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

func resourceGroupBinding() *schema.Resource {
	return &schema.Resource{
		Description: "This resource adds a single looker group to a parent group. If the fields in this resource are modified, it will be destroyed and recreated.",

		// no update method needed as resource is destroyed and recreated if any fields are modified
		CreateContext: resourceGroupBindingCreate,
		ReadContext:   resourceGroupBindingRead,
		DeleteContext: resourceGroupBindingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupBindingImport,
		},

		Schema: map[string]*schema.Schema{
			"parent_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the parent group",
				ForceNew:    true,
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the group to be added to the parent group",
				ForceNew:    true,
			},
		},
	}
}

func resourceGroupBindingCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	parentGroupID := d.Get("parent_group_id").(string)
	groupID := d.Get("group_id").(string)

	_, groupErr := api.AddGroupGroup(parentGroupID,
		sdk.GroupIdForGroupInclusion{
			GroupId: conv.PString(groupID),
		},
		nil,
	)
	if groupErr != nil {
		return diag.FromErr(groupErr)
	}

	d.SetId(fmt.Sprintf("%s_%s", parentGroupID, groupID))

	return resourceGroupBindingRead(ctx, d, c)
}

func resourceGroupBindingRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groups, grErr := api.SearchGroupsWithHierarchy(sdk.RequestSearchGroups{
		Id: conv.PString(d.Get("group_id").(string)),
	}, nil)
	if grErr != nil {
		return diag.FromErr(grErr)
	}

	// search result should only have one entry per group_id
	// this case is almost impossible because every group is, by default, binded to the default 'All Users' looker group.
	if len(groups) != 1 {
		d.SetId("")
		return nil
	}

	// inspect parent groups and check if parent group is contained in this slice
	if groups[0].ParentGroupIds != nil && !slice.Contains(*groups[0].ParentGroupIds, d.Get("parent_group_id").(string)) {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceGroupBindingDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	delErr := api.DeleteGroupFromGroup(
		d.Get("parent_group_id").(string),
		d.Get("group_id").(string),
		nil,
	)
	if delErr != nil && !errors.Is(delErr, io.EOF) {
		return diag.FromErr(delErr)
	}

	return nil
}

func resourceGroupBindingImport(ctx context.Context, d *schema.ResourceData, c interface{}) ([]*schema.ResourceData, error) {
	// id is <parent_group_id>_<group_id>
	s := strings.Split(d.Id(), "_")
	if len(s) < 2 {
		diag.Errorf("invalid id, should be of the form <parent_group_id>_<group_id>")
	}

	resErr := multierror.Append(
		d.Set("parent_group_id", s[0]),
		d.Set("group_id", s[1]),
	).ErrorOrNil()
	if resErr != nil {
		return nil, resErr
	}

	return []*schema.ResourceData{d}, nil
}
