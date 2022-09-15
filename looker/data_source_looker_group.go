package looker

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func datasourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "This datasource reads a looker group from a Looker instance.",

		ReadContext: datasourceGroupRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "The name of the group. This field is case sensitive.",
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "The id of the group",
			},
		},
	}
}

func datasourceGroupRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// exactly one of these variables will be nil - this is enforced by the data source schema
	name := conv.PString(d.Get("name").(string))
	id := conv.PString(d.Get("id").(string))

	groups, grErr := api.SearchGroups(sdk.RequestSearchGroups{
		Name: name,
		Id:   id,
	}, nil)
	if grErr != nil {
		return diag.FromErr(grErr)
	}

	var group *sdk.Group
	for _, gr := range groups {
		if id != nil && gr.Id != nil && *id == *gr.Id {
			group = &gr
			break
		}

		if name != nil && gr.Name != nil && *name == *gr.Name {
			group = &gr
			break
		}
	}

	if group == nil {
		if id != nil {
			return diag.Errorf("group with id %s not found", *id)
		}
		return diag.Errorf("group with name '%s' not found", *name)
	}
	d.SetId(*group.Id)

	if group.Name == nil {
		return diag.Errorf("group name is missing for group with id %s", *group.Id)
	}
	return diag.FromErr(d.Set("name", group.Name))
}
