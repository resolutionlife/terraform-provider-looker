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
)

func resourceUserAttributeUser() *schema.Resource {
	return &schema.Resource{
		Description: "This resource sets a value onto a user for the given user attribute. If a default value is already set for the user attribute, this value will override the default value. Note that if the user attribute values are hidden (can be configured when provisioning a `looker_user_attribute`) then the provider does not have the permissions to read the hidden values, and cannot verify if the value has been manually changed in the Looker UI. The provider can however check if the value has been removed, and will prompt to recreate the resource.",

		CreateContext: resourceUserAttributeUserCreate,
		ReadContext:   resourceUserAttributeUserRead,
		UpdateContext: resourceUserAttributeUserCreate,
		DeleteContext: resourceUserAttributeUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserAttributeUserImport,
		},

		Schema: map[string]*schema.Schema{
			"user_attribute_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user attribute",
				ForceNew:    true,
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user",
				ForceNew:    true,
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value of the user attribute to be set on the user",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the resource. This id is of the form <user_attribute_id>_<user_id>",
			},
		},
	}
}

func resourceUserAttributeUserCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttrID := d.Get("user_attribute_id").(string)
	userID := d.Get("user_id").(string)

	_, setErr := api.SetUserAttributeUserValue(userID, userAttrID,
		sdk.WriteUserAttributeWithValue{
			Value: conv.PString(d.Get("value").(string)),
		},
		nil,
	)
	if setErr != nil {
		return diag.FromErr(setErr)
	}
	d.SetId(fmt.Sprintf("%s_%s", userAttrID, userID))

	return resourceUserAttributeUserRead(ctx, d, c)
}

func resourceUserAttributeUserRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userID := d.Get("user_id").(string)

	// the UserAttributeIds field in sdk.RequestUserAttributeUserValues should be type *string but is of type *rtl.DelimString.
	// TODO: amend this request to search for a specific user_attribute_id when the SDK is fixed.
	userAttrs, uaErr := api.UserAttributeUserValues(sdk.RequestUserAttributeUserValues{
		UserId: userID,
		Fields: conv.PString("user_attribute_id,user_id,value,value_is_hidden,source"),
	}, nil)
	if uaErr != nil {
		return diag.FromErr(uaErr)
	}

	for _, ua := range userAttrs {
		if ua.UserAttributeId != nil && *ua.UserAttributeId == d.Get("user_attribute_id").(string) {

			var setValueErr error
			// if the value is hidden then the api does not allow any user to read the value to check if it has changed
			// we can check that a value not been deleted, as the attribute source will either revert to "Default" or "No Value" if no default value is set on the user attribute
			if ua.ValueIsHidden != nil && *ua.ValueIsHidden {
				if ua.Source != nil && *ua.Source == "Default" || *ua.Source == "No Value" {
					d.SetId("")
					return nil
				}
			} else {
				setValueErr = d.Set("value", ua.Value)
			}

			// if the user attribute value is not hidden then set the value in the state
			result := multierror.Append(
				setValueErr,
				d.Set("user_attribute_id", ua.UserAttributeId),
				d.Set("user_id", ua.UserId),
			)

			return diag.FromErr(result.ErrorOrNil())
		}
	}

	// user attribute is not found on the user
	d.SetId("")
	return nil
}

func resourceUserAttributeUserDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttrID := d.Get("user_attribute_id").(string)
	userID := d.Get("user_id").(string)

	// this delete removes the custom value and reverts back to the default value of the attribute, if set.
	delErr := api.DeleteUserAttributeUserValue(userID, userAttrID, nil)
	if delErr != nil && !errors.Is(delErr, io.EOF) {
		// TODO: cover the case when the user has already been deleted
		return diag.FromErr(delErr)
	}

	return nil
}

func resourceUserAttributeUserImport(ctx context.Context, d *schema.ResourceData, c interface{}) ([]*schema.ResourceData, error) {
	// id is <user_attribute_id>_<user_id>
	s := strings.Split(d.Id(), "_")
	if len(s) < 2 {
		diag.Errorf("invalid id, should be of the form <parent_group_id>_<group_id>")
	}

	resErr := multierror.Append(
		d.Set("user_attribute_id", s[0]),
		d.Set("user_id", s[1]),
	).ErrorOrNil()
	if resErr != nil {
		return nil, resErr
	}

	return []*schema.ResourceData{d}, nil
}
