package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func dataSourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		Description: "This datasource reads a permission set from a Looker instance.",

		ReadContext: dataSourcePermissionSetRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the permission set. This field is case sensitive. Documentation for default permission sets can be found [here](https://docs.looker.com/admin-options/settings/roles#permission_sets).",
				ExactlyOneOf: []string{"name", "id"},
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The id of the permission set.",
				ExactlyOneOf: []string{"name", "id"},
			},
			"permissions": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "A list of permissions within the permission set.",
			},
		},
	}
}

func dataSourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// exactly one of these variables will be nil - this is enforced by the data source schema
	name := conv.PString(d.Get("name").(string))
	id := conv.PString(d.Get("id").(string))

	permSets, permSetsErr := api.SearchPermissionSets(
		sdk.RequestSearchModelSets{
			Name: name,
			Id:   id,
		}, nil,
	)
	if permSetsErr != nil {
		return diag.FromErr(permSetsErr)
	}

	var ps *sdk.PermissionSet
	for _, p := range permSets {
		// if id is supplied, search for matching id
		if id != nil && p.Id != nil && *p.Id == *id {
			ps = &p
			break
		}

		// if name is supplied, search for matching name
		if name != nil && p.Name != nil && *p.Name == *name {
			ps = &p
			break
		}
	}
	if ps == nil {
		return diag.Errorf("no permission set found")
	}
	// response id is always populated
	d.SetId(*ps.Id)

	if ps.Name == nil {
		return diag.Errorf("name not found for permission set with id: %s", *ps.Id)
	}
	result := multierror.Append(
		d.Set("name", ps.Name),
		d.Set("permissions", ps.Permissions),
	)

	return diag.FromErr(result.ErrorOrNil())
}
