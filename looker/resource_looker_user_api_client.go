package looker

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func resourceUserAPIClient() *schema.Resource {
	return &schema.Resource{
		Description: "This resource creates API credentials for a user.",

		CreateContext: resourceUserAPIClientCreate,
		ReadContext:   resourceUserAPIClientRead,
		DeleteContext: resourceUserAPIClientDelete,

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeString,
				Description: "The ID of the user.",
				Required:    true,
				ForceNew:    true,
			},
			"client_id": {
				Type:        schema.TypeString,
				Description: "The ID of the client.",
				Computed:    true,
			},
			"client_secret": {
				Type:        schema.TypeString,
				Description: "The secret for the client.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceUserAPIClientCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	res, err := api.CreateUserCredentialsApi3(d.Get("user_id").(string), "", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	if res.ClientId == nil {
		return diag.Errorf("client_id is missing")
	}
	if err := d.Set("client_id", *res.ClientId); err != nil {
		return diag.FromErr(err)
	}

	if res.ClientSecret == nil {
		return diag.Errorf("client_secret is missing")
	}
	if err := d.Set("client_secret", *res.ClientSecret); err != nil {
		return diag.FromErr(err)
	}

	if res.Id == nil {
		return diag.Errorf("API credentials ID is missing")
	}
	d.SetId(*res.Id)

	return resourceUserAPIClientRead(ctx, d, c)
}

func resourceUserAPIClientRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// TODO: handle case when API credentials are not found.
	_, err := api.UserCredentialsApi3(d.Get("user_id").(string), d.Id(), "", nil)
	return diag.FromErr(err)
}

func resourceUserAPIClientDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// TODO: handle case when API credentials are not found.
	if _, err := api.DeleteUserCredentialsApi3(d.Get("user_id").(string), d.Id(), nil); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
