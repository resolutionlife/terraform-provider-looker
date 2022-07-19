package looker

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func dataSourcePermissionSet() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePermissionSetRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the permission set. Documentation for default permission sets can be found [here](https://docs.looker.com/admin-options/settings/roles#permission_sets)",
			},
			"id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The id of the resource",
			},
		},
	}
}

func dataSourcePermissionSetRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	permSets, permSetsErr := api.AllPermissionSets("id, name", nil)
	if permSetsErr != nil {
		return diag.FromErr(permSetsErr)
	}

	permSetName := d.Get("name").(string)
	for _, set := range permSets {
		if set.Name != nil && *set.Name == permSetName {
			if set.Id == nil {
				return diag.Errorf("permission set %s has nil id", permSetName)
			}
			d.SetId(*set.Id)
			return nil
		}
	}

	return diag.Errorf("no permission set found with the name %s", permSetName)
}
