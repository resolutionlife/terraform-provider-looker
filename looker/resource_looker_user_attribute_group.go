package looker

import (
	"context"
	"errors"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceUserAttributeGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages Looker User Attribute Groups",
		CreateContext: resourceUserAttributeGroupCreate,
		ReadContext:   resourceUserAttributeGroupRead,
		UpdateContext: resourceUserAttributeGroupUpdate,
		DeleteContext: resourceUserAttributeGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique id of the user attribute group",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the group to set user attribute for",
			},
			"user_attribute_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user attribute to assign",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of attribute overriding any existing default value",
			},
		},
	}
}

func resourceUserAttributeGroupCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	usrAttrGrp, err := api.UpdateUserAttributeGroupValue(
		d.Get("group_id").(string),
		d.Get("user_attribute_id").(string),
		sdk.UserAttributeGroupValue{
			GroupId:         conv.PString(d.Get("group_id").(string)),
			UserAttributeId: conv.PString(d.Get("user_attribute_id").(string)),
			Value:           conv.PString(d.Get("value").(string)),
		},
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if usrAttrGrp.Id == nil {
		return diag.Errorf("user attribute group has missing id")
	}
	d.SetId(*usrAttrGrp.Id)

	tflog.Info(ctx, "CREATE", map[string]interface{}{
		"groupId": *usrAttrGrp.GroupId,
		"attrId":  *usrAttrGrp.UserAttributeId,
		"value":   *usrAttrGrp.Value,
	})

	return resourceUserAttributeGroupRead(ctx, d, c)
}

func resourceUserAttributeGroupRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	usrAttrGrps, err := api.AllUserAttributeGroupValues(
		d.Get("user_attribute_id").(string),
		"",
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	usrAttrGrp := getAttributeByGroupId(usrAttrGrps, d.Get("group_id").(string))
	if usrAttrGrp == nil {
		return diag.Errorf("unable to match user attribute to group")
	}

	value := usrAttrGrp.Value
	// if hidden, default value cannot be retrieved from API
	if *usrAttrGrp.ValueIsHidden {
		value = conv.PString(d.Get("value").(string))
	}

	result := multierror.Append(
		d.Set("id", usrAttrGrp.Id),
		d.Set("group_id", usrAttrGrp.GroupId),
		d.Set("user_attribute_id", usrAttrGrp.UserAttributeId),
		d.Set("value", value),
	)

	tflog.Info(ctx, "READ", map[string]interface{}{
		"groupId": usrAttrGrp.GroupId,
		"attrId":  usrAttrGrp.UserAttributeId,
		"value":   value,
	})

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserAttributeGroupUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	usrAttrGrp, err := api.UpdateUserAttributeGroupValue(
		d.Get("group_id").(string),
		d.Get("user_attribute_id").(string),
		sdk.UserAttributeGroupValue{
			GroupId:         conv.PString(d.Get("group_id").(string)),
			UserAttributeId: conv.PString(d.Get("user_attribute_id").(string)),
			Value:           conv.PString(d.Get("value").(string)),
		},
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Info(ctx, "UPDATE", map[string]interface{}{
		"groupId": *usrAttrGrp.GroupId,
		"attrId":  *usrAttrGrp.UserAttributeId,
		"value":   *usrAttrGrp.Value,
	})

	return resourceUserAttributeGroupRead(ctx, d, c)
}

func resourceUserAttributeGroupDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	err := api.DeleteUserAttributeGroupValue(
		d.Get("group_id").(string),
		d.Get("user_attribute_id").(string),
		nil,
	)
	if err != nil && !errors.Is(err, io.EOF) {
		return diag.FromErr(err)
	}

	return nil
}

func getAttributeByGroupId(usrAttrGrps []sdk.UserAttributeGroupValue, grpId string) *sdk.UserAttributeGroupValue {
	for _, usrAttrGrp := range usrAttrGrps {
		if *usrAttrGrp.GroupId == grpId {
			return &usrAttrGrp
		}
	}

	return nil
}
