package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Manages users of a Looker instance",

		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the group",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the group",
			},
			"externally_managed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether membership to this group is managed outside of looker",
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	group, grErr := api.CreateGroup(
		sdk.WriteGroup{
			Name: conv.PString(d.Get("name").(string)),
		},
		"id,name", nil,
	)
	if grErr != nil {
		return diag.FromErr(grErr)
	}

	if group.Id == nil {
		return diag.Errorf("group %s has missing id", group.Name)
	}
	d.SetId(*group.Id)

	return resourceGroupRead(ctx, d, c)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	group, grErr := api.Group(d.Id(), "id,name,externally_managed", nil)
	if grErr != nil {
		// TODO: handle the case where a user is not found
		return diag.FromErr(grErr)
	}

	result := multierror.Append(
		d.Set("name", group.Name),
		d.Set("id", group.Id),
		d.Set("externally_managed", group.ExternallyManaged),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	if !d.HasChange("name") {
		return nil
	}

	_, grErr := api.UpdateGroup(d.Id(),
		sdk.WriteGroup{
			Name: conv.PString(d.Get("name").(string)),
		},
		"", nil,
	)
	if grErr != nil {
		// TODO: handle the case where a user is not found
		return diag.FromErr(grErr)
	}
	return resourceGroupRead(ctx, d, c)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	_, delErr := api.DeleteGroup(d.Id(), nil)
	if delErr != nil {
		// TODO: handle the case where a user is not found
		return diag.FromErr(delErr)
	}

	return nil
}
