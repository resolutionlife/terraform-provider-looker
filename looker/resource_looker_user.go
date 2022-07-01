package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

var userResource = schema.Resource{
	Description: "Manages users of a Looker instance",

	Schema: map[string]*schema.Schema{
		"email": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The email address of the user",
		},
		"first_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The first name of the user",
		},
		"last_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The last name of the user",
		},
	},

	CreateContext: resourceUserCreate,
	ReadContext:   resourceUserRead,
	UpdateContext: resourceUserUpdate,
	DeleteContext: resourceUserDelete,
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	user, userErr := api.CreateUser(
		sdk.WriteUser{
			FirstName: pString(d.Get("first_name").(string)),
			LastName:  pString(d.Get("last_name").(string)),
		}, "", nil,
	)
	if userErr != nil {
		return diag.FromErr(userErr)
	}

	if user.Id == nil {
		return diag.Errorf("user id not set")
	}
	d.SetId(*user.Id)

	_, credErr := api.CreateUserCredentialsEmail(*user.Id,
		sdk.WriteCredentialsEmail{
			Email: pString(d.Get("email").(string)),
		}, "", nil,
	)
	if credErr != nil {
		return diag.FromErr(credErr)
	}

	_, sendEmailErr := api.SendUserCredentialsEmailPasswordReset(*user.Id, "", nil)
	if sendEmailErr != nil {
		return diag.FromErr(sendEmailErr)
	}

	return resourceUserRead(ctx, d, c)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Id()
	user, userErr := api.User(userID, "", nil)
	if userErr != nil {
		d.SetId("")
		return diag.FromErr(userErr)
	}

	result := multierror.Append(
		d.Set("email", user.Email),
		d.Set("first_name", user.FirstName),
		d.Set("last_name", user.LastName),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Id()
	if d.HasChanges("first_name", "last_name") {
		_, updateErr := api.UpdateUser(userID,
			sdk.WriteUser{
				FirstName: pString(d.Get("first_name").(string)),
				LastName:  pString(d.Get("last_name").(string)),
			}, "", nil,
		)
		if updateErr != nil {
			return diag.FromErr(updateErr)
		}
	}

	if d.HasChange("email") {
		_, updateCredsErr := api.UpdateUserCredentialsEmail(userID,
			sdk.WriteCredentialsEmail{
				Email: pString(d.Get("email").(string)),
			}, "", nil,
		)
		if updateCredsErr != nil {
			return diag.FromErr(updateCredsErr)
		}
	}

	return resourceUserRead(ctx, d, c)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	_, delErr := api.DeleteUser(d.Id(), nil)
	if delErr != nil {
		d.SetId("")
		return diag.FromErr(delErr)
	}

	return nil
}

func pString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
