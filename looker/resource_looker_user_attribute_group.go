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
			"user_attribute_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The id of the user attribute to assign",
				ForceNew:    true,
			},
			"group_values": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The id of the group to set user attribute for",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value of attribute overriding any existing default value",
						},
					},
				},
			},
		},
	}
}

func resourceUserAttributeGroupCreate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupValues := d.Get("group_values").([]interface{})

	var userAttrGroupVaules []sdk.UserAttributeGroupValue
	for _, groupValue := range groupValues {
		groupValueMap := groupValue.(map[string]interface{})
		userAttrGroupVaules = append(userAttrGroupVaules, sdk.UserAttributeGroupValue{
			GroupId: conv.PString(groupValueMap["group_id"].(string)),
			Value:   conv.PString(groupValueMap["value"].(string)),
		})
	}

	userAttrGroups, err := api.SetUserAttributeGroupValues(
		d.Get("user_attribute_id").(string),
		userAttrGroupVaules,
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var userAttrGroupId string
	for _, userAttrGroup := range userAttrGroups {
		if userAttrGroup.Id == nil {
			return diag.Errorf("user attribute group has missing id")
		}
		userAttrGroupId += *userAttrGroup.Id

		tflog.Info(ctx, "CREATE", map[string]interface{}{
			"groupId": *userAttrGroup.GroupId,
			"attrId":  *userAttrGroup.UserAttributeId,
			"value":   *userAttrGroup.Value,
		})
	}
	d.SetId(userAttrGroupId)

	return resourceUserAttributeGroupRead(ctx, d, c)
}

func resourceUserAttributeGroupRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	userAttrGroups, err := api.AllUserAttributeGroupValues(
		d.Get("user_attribute_id").(string),
		"",
		nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("user_attribute_id", userAttrGroups[0].UserAttributeId), // all share the same user attribute, therefore can just the first
		d.Set("group_values", buildGroupValuesMap(ctx, d, userAttrGroups)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func resourceUserAttributeGroupUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupValues := d.Get("group_values").([]interface{})
	for _, groupValue := range groupValues {
		groupValueMap := groupValue.(map[string]interface{})
		usrAttrGrp, err := api.UpdateUserAttributeGroupValue(
			groupValueMap["group_id"].(string),
			d.Get("user_attribute_id").(string),
			sdk.UserAttributeGroupValue{
				GroupId:         conv.PString(groupValueMap["group_id"].(string)),
				UserAttributeId: conv.PString(d.Get("user_attribute_id").(string)),
				Value:           conv.PString(groupValueMap["value"].(string)),
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
	}

	return resourceUserAttributeGroupRead(ctx, d, c)
}

func resourceUserAttributeGroupDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupValues := d.Get("group_values").([]interface{})
	for _, groupValue := range groupValues {
		groupValueMap := groupValue.(map[string]interface{})
		err := api.DeleteUserAttributeGroupValue(
			groupValueMap["group_id"].(string),
			d.Get("user_attribute_id").(string),
			nil,
		)
		if err != nil && !errors.Is(err, io.EOF) {
			return diag.FromErr(err)
		}
	}

	return nil
}

func buildGroupValuesMap(ctx context.Context, d *schema.ResourceData, userAttrGroups []sdk.UserAttributeGroupValue) []map[string]interface{} {
	var groupValuesMap []map[string]interface{}

	groupValues := d.Get("group_values").([]interface{})
	for i, userAttrGroup := range userAttrGroups {
		groupid := *userAttrGroup.GroupId
		value := *userAttrGroup.Value

		if *userAttrGroup.ValueIsHidden {
			value = groupValues[i].(map[string]interface{})["value"].(string)
		}

		groupValuesMap = append(groupValuesMap, map[string]interface{}{
			"group_id": groupid,
			"value":    value,
		})

		tflog.Info(ctx, "READ", map[string]interface{}{
			"groupId": *userAttrGroup.GroupId,
			"attrId":  *userAttrGroup.UserAttributeId,
			"value":   *userAttrGroup.Value,
		})
	}

	return groupValuesMap
}
