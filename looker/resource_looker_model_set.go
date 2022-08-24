package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceModelSet() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a set of Looker models",
		CreateContext: resourceModelSetCreate,
		ReadContext:   resourceModelSetRead,
		UpdateContext: resourceModelSetUpdate,
		DeleteContext: resourceModelSetDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of the model set",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the model set",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the model set",
			},
			"models": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "The list of models in the model set",
			},
		},
	}
}

func resourceModelSetCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	models, ok := d.Get("models").(*schema.Set)
	if !ok {
		return diag.Errorf("models is not a set")
	}

	modelsSlice, err := conv.SchemaSetToSliceString(models)
	if err != nil {
		return diag.FromErr(err)
	}

	modelSet, err := api.CreateModelSet(sdk.WriteModelSet{
		Name:   conv.PString(d.Get("name").(string)),
		Models: conv.PSlices(modelsSlice),
	}, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	if modelSet.Id == nil {
		return diag.Errorf("model set has missing id")
	}
	d.SetId(*modelSet.Id)

	return resourceModelSetRead(ctx, d, c)
}

func resourceModelSetRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	modelSet, err := api.ModelSet(d.Id(), "id,name,models", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	if modelSet.Id != nil {
		return diag.Errorf("")
	}

	d.SetId(*modelSet.Id)
	result := multierror.Append(
		d.Set("name", modelSet.Name),
		d.Set("models", modelSet.Models),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceModelSetUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	models, ok := d.Get("models").(*schema.Set)
	if !ok {
		return diag.Errorf("models is not a set")
	}

	modelsSlice, err := conv.SchemaSetToSliceString(models)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.UpdateModelSet(
		d.Id(),
		sdk.WriteModelSet{
			Name:   conv.PString(d.Get("name").(string)),
			Models: conv.PSlices(modelsSlice),
		},
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceModelSetRead(ctx, d, c)
}

func resourceModelSetDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	_, err := api.DeleteModelSet(d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
