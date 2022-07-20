package looker

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func datasourceModelSet() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceModelSetRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the model set. This field is not case sensitive. Documentation on model sets can be found [here](https://docs.looker.com/admin-options/settings/roles#model_sets).",
			},
			"id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The id of the resource",
			},
		},
	}
}

func datasourceModelSetRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	modelSets, modelSetErr := api.AllModelSets("id, name", nil)
	if modelSetErr != nil {
		return diag.FromErr(modelSetErr)
	}

	setName := d.Get("name").(string)
	for _, set := range modelSets {
		if set.Name != nil && strings.EqualFold(*set.Name, setName) {
			if set.Id == nil {
				return diag.Errorf("model set %s has nil id", setName)
			}
			d.SetId(*set.Id)
			return nil
		}
	}

	return diag.Errorf("no model set found with the name %s", setName)
}
