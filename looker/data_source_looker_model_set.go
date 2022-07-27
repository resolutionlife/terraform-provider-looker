package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func datasourceModelSet() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceModelSetRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the model set. This field is case sensitive. Documentation on model sets can be found [here](https://docs.looker.com/admin-options/settings/roles#model_sets).",
				ExactlyOneOf: []string{"name", "id"},
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The id of the resource",
				ExactlyOneOf: []string{"name", "id"},
			},
			"models": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "A list of models within the model set.",
			},
		},
	}
}

func dataSourceModelSetRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	name := conv.PString(d.Get("name").(string))
	id := conv.PString(d.Get("id").(string))

	modelSets, modelSetsErr := api.SearchModelSets(
		sdk.RequestSearchModelSets{
			Name: name,
			Id:   id,
		}, nil,
	)
	if modelSetsErr != nil {
		return diag.FromErr(modelSetsErr)
	}

	var ms *sdk.ModelSet
	for _, m := range modelSets {
		// if id is supplied, search for matching id
		if id != nil && m.Id != nil && *m.Id == *id {
			ms = &m
			break
		}

		// if name is supplied, search for matching name
		if name != nil && m.Name != nil && *m.Name == *name {
			ms = &m
			break
		}
	}
	if ms == nil {
		return diag.Errorf("no model set found")
	}
	// response id is always populated
	d.SetId(*ms.Id)

	if ms.Name == nil {
		return diag.Errorf("name not found for model set with id: %s", *ms.Id)
	}
	result := multierror.Append(
		d.Set("name", ms.Name),
		d.Set("models", ms.Models),
	)

	return diag.FromErr(result.ErrorOrNil())
}
