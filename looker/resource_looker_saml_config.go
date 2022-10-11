package looker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func resourceSamlConfig() *schema.Resource {
	return &schema.Resource{
		Description: "This resource updates the SAML config in a Looker instance.",

		CreateContext: resourceSamlConfigCreateOrUpdate,
		ReadContext:   resourceSamlConfigRead,
		UpdateContext: resourceSamlConfigCreateOrUpdate,
		DeleteContext: resourceSamlConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Allows roles to be directly assigned to SAML auth'd users.",
			},
			"idp_cert": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identity Provider Certificate (provided by IdP)",
				Sensitive:   true,
			},
			"idp_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identity Provider Url (provided by IdP)",
			},
			"idp_issuer": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identity Provider Issuer (provided by IdP)",
			},
			"idp_audience": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identity Provider Audience (set in IdP config). Optional in Looker. Set this only if you want Looker to validate the audience value returned by the IdP.",
			},
			"allowed_clock_drift": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Count of seconds of clock drift to allow when validating timestamps of assertions.",
			},
			"user_attribute_map_email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of user record attributes used to indicate email address field",
			},
			"user_attribute_map_first_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of user record attributes used to indicate first name",
			},
			"user_attribute_map_last_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of user record attributes used to indicate last name",
			},
			"new_user_migration_types": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Merge first-time saml login to existing user account by email addresses. When a user logs in for the first time via saml this option will connect this user into their existing account by finding the account with a matching email address by testing the given types of credentials for existing users. Otherwise a new user account will be created for the user.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
						"email", "ldap", "google",
					}, false)),
				},
			},
			"default_new_user_role_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Array of ids of roles that will be applied to new users the first time they login via Saml",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_new_user_group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Array of ids of groups that will be applied to new users the first time they login via Saml",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auth_requires_role": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Users will not be allowed to login at all unless a role for them is found in Saml if set to true",
			},
			"bypass_login_page": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Bypass the login page when user authentication is required. Redirect to IdP immediately instead.",
			},
			"allow_direct_roles": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Allows roles to be directly assigned to SAML auth'd users.",
			},
			"allow_normal_group_membership": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Allow SAML auth'd users to be members of non-reflected Looker groups. If 'false', user will be removed from non-reflected groups on login.",
			},
			"alternate_email_login_allowed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Allow alternate email-based login via '/login/email' for admins and for specified users with the 'login_special_email' permission. This option is useful as a fallback during ldap setup, if ldap config problems occur later, or if you need to support some users who are not in your ldap directory. Looker email/password logins are always disabled for regular users when ldap is enabled.",
			},
			"allow_roles_from_normal_groups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "SAML auth'd users will inherit roles from non-reflected Looker groups.",
			},
			"groups_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of user record attributes used to indicate groups. Used when 'groups_finder_type' is set to 'grouped_attribute_values'",
			},
			"groups_finder_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "grouped_attribute_values",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"grouped_attribute_values",
					"individual_attributes",
				}, false)),
				Description: "Identifier for a strategy for how Looker will find groups in the SAML response.",
			},
			"groups_with_role_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Array of mappings between Saml Groups and arrays of Looker Role ids",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique Id",
						},
						"looker_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Unique Id of group in Looker",
						},
						"looker_group_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of group in Looker",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of group in Saml",
						},
						"role_ids": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Looker Role Ids",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"groups_member_value": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Value for group attribute used to indicate membership. Used when 'groups_finder_type' is set to 'individual_attributes'",
			},
			"set_roles_from_groups": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set user roles in Looker based on groups from Saml",
			},
			"user_attributes_with_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Array of mappings between Saml User Attributes and arrays of Looker User Attribute ids",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of User Attribute in Saml",
						},
						"required": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Required to be in Saml assertion for login to be allowed to succeed",
						},
						"user_attribute_ids": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Looker User Attribute Ids",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"id": {
				Description: "This is set to a random value at create time",
				Computed:    true,
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceSamlConfigRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	cfg, err := api.SamlConfig(nil)
	if errors.Is(err, sdk.ErrNotFound) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("enabled", cfg.Enabled),
		d.Set("idp_cert", cfg.IdpCert),
		d.Set("idp_url", cfg.IdpUrl),
		d.Set("idp_issuer", cfg.IdpIssuer),
		d.Set("idp_audience", cfg.IdpAudience),
		d.Set("allowed_clock_drift", cfg.AllowedClockDrift),
		d.Set("user_attribute_map_email", cfg.UserAttributeMapEmail),
		d.Set("user_attribute_map_first_name", cfg.UserAttributeMapFirstName),
		d.Set("user_attribute_map_last_name", cfg.UserAttributeMapLastName),
		d.Set("default_new_user_role_ids", flattenRoles(cfg.DefaultNewUserRoles)),
		d.Set("default_new_user_group_ids", flattenGroups(cfg.DefaultNewUserGroups)),
		d.Set("auth_requires_role", cfg.AuthRequiresRole),
		d.Set("bypass_login_page", cfg.BypassLoginPage),
		d.Set("allow_direct_roles", cfg.AllowDirectRoles),
		d.Set("allow_normal_group_membership", cfg.AllowNormalGroupMembership),
		d.Set("alternate_email_login_allowed", cfg.AlternateEmailLoginAllowed),
		d.Set("allow_roles_from_normal_groups", cfg.AllowRolesFromNormalGroups),
		d.Set("groups_attribute", cfg.GroupsAttribute),
		d.Set("groups_finder_type", cfg.GroupsFinderType),
		d.Set("groups_with_role_ids", flattenGroupsWithRoleIDs(cfg.Groups)),
		d.Set("groups_member_value", cfg.GroupsMemberValue),
		d.Set("set_roles_from_groups", cfg.SetRolesFromGroups),
		d.Set("user_attributes_with_ids", flattenUserAttributesWithIDs(cfg.UserAttributes)),
	)

	if cfg.NewUserMigrationTypes != nil {
		result = multierror.Append(result, d.Set("new_user_migration_types", strings.Split(*cfg.NewUserMigrationTypes, ",")))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func resourceSamlConfigCreateOrUpdate(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	groupsWithRoleIDs := make([]sdk.SamlGroupWrite, 0)
	if vs, ok := d.GetOk("groups_with_role_ids"); ok {
		for _, v := range vs.(*schema.Set).List() {
			sgw := v.(map[string]interface{})
			roleIDs, err := conv.SchemaSetToSliceString(sgw["role_ids"].(*schema.Set))
			if err != nil {
				return diag.FromErr(err)
			}

			write := sdk.SamlGroupWrite{
				LookerGroupName: conv.P(sgw["looker_group_name"].(string)),
				Name:            conv.P(sgw["name"].(string)),
				RoleIds:         conv.P(roleIDs),
			}

			// check for existing ID to ensure we update existing relationship
			if _, ok := sgw["id"]; ok && sgw["id"].(string) != "" {
				write.Id = conv.P(sgw["id"].(string))
			}

			groupsWithRoleIDs = append(groupsWithRoleIDs, write)
		}
	}

	userAttributesWithIds := make([]sdk.SamlUserAttributeWrite, 0)
	if vs, ok := d.GetOk("user_attributes_with_ids"); ok {
		for _, v := range vs.(*schema.Set).List() {
			suaw := v.(map[string]interface{})
			userAttributeIDs, err := conv.SchemaSetToSliceString(suaw["user_attribute_ids"].(*schema.Set))
			if err != nil {
				return diag.FromErr(err)
			}
			userAttributesWithIds = append(userAttributesWithIds, sdk.SamlUserAttributeWrite{
				Name:             conv.P(suaw["name"].(string)),
				Required:         conv.P(suaw["required"].(bool)),
				UserAttributeIds: conv.P(userAttributeIDs),
			})
		}
	}

	newUserMigrationTypes, err := conv.SchemaSetToSliceString(d.Get("new_user_migration_types").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	defaultNewUserGroupIds, err := conv.SchemaSetToSliceString(d.Get("default_new_user_group_ids").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	defaultNewUserRoleIds, err := conv.SchemaSetToSliceString(d.Get("default_new_user_role_ids").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}

	cfg := sdk.WriteSamlConfig{
		Enabled:                    conv.P(d.Get("enabled").(bool)),
		IdpCert:                    conv.P(d.Get("idp_cert").(string)),
		IdpUrl:                     conv.P(d.Get("idp_url").(string)),
		IdpIssuer:                  conv.P(d.Get("idp_issuer").(string)),
		IdpAudience:                conv.P(d.Get("idp_audience").(string)),
		AllowedClockDrift:          conv.P(int64(d.Get("allowed_clock_drift").(int))),
		UserAttributeMapEmail:      conv.P(d.Get("user_attribute_map_email").(string)),
		UserAttributeMapFirstName:  conv.P(d.Get("user_attribute_map_first_name").(string)),
		UserAttributeMapLastName:   conv.P(d.Get("user_attribute_map_last_name").(string)),
		NewUserMigrationTypes:      conv.P(strings.Join(newUserMigrationTypes, ",")),
		DefaultNewUserGroupIds:     conv.P(defaultNewUserGroupIds),
		DefaultNewUserRoleIds:      conv.P(defaultNewUserRoleIds),
		AuthRequiresRole:           conv.P(d.Get("auth_requires_role").(bool)),
		BypassLoginPage:            conv.P(d.Get("bypass_login_page").(bool)),
		AllowDirectRoles:           conv.P(d.Get("allow_direct_roles").(bool)),
		AllowNormalGroupMembership: conv.P(d.Get("allow_normal_group_membership").(bool)),
		AlternateEmailLoginAllowed: conv.P(d.Get("alternate_email_login_allowed").(bool)),
		AllowRolesFromNormalGroups: conv.P(d.Get("allow_roles_from_normal_groups").(bool)),
		GroupsAttribute:            conv.P(d.Get("groups_attribute").(string)),
		GroupsFinderType:           conv.P(d.Get("groups_finder_type").(string)),
		GroupsWithRoleIds:          conv.P(groupsWithRoleIDs),
		GroupsMemberValue:          conv.P(d.Get("groups_member_value").(string)),
		SetRolesFromGroups:         conv.P(d.Get("set_roles_from_groups").(bool)),
		UserAttributesWithIds:      conv.P(userAttributesWithIds),
	}

	if _, err := api.UpdateSamlConfig(cfg, nil); err != nil {
		if errors.Is(err, sdk.ErrNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", rand.Int()))

	return resourceSamlConfigRead(ctx, d, c)
}

func resourceSamlConfigDelete(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	if _, err := api.UpdateSamlConfig(sdk.WriteSamlConfig{
		Enabled:                    conv.P(false),
		IdpCert:                    conv.P(""),
		IdpUrl:                     conv.P(""),
		IdpIssuer:                  conv.P(""),
		IdpAudience:                conv.P(""),
		AllowedClockDrift:          conv.P(int64(0)),
		UserAttributeMapEmail:      conv.P(""),
		UserAttributeMapFirstName:  conv.P(""),
		UserAttributeMapLastName:   conv.P(""),
		NewUserMigrationTypes:      conv.P(""),
		DefaultNewUserGroupIds:     conv.P([]string{}),
		DefaultNewUserRoleIds:      conv.P([]string{}),
		AuthRequiresRole:           conv.P(false),
		BypassLoginPage:            conv.P(false),
		AllowDirectRoles:           conv.P(false),
		AllowNormalGroupMembership: conv.P(false),
		AlternateEmailLoginAllowed: conv.P(false),
		AllowRolesFromNormalGroups: conv.P(false),
		GroupsAttribute:            conv.P(""),
		GroupsFinderType:           conv.P(""),
		GroupsWithRoleIds:          conv.P([]sdk.SamlGroupWrite{}),
		GroupsMemberValue:          conv.P(""),
		SetRolesFromGroups:         conv.P(false),
		UserAttributesWithIds:      conv.P([]sdk.SamlUserAttributeWrite{}),
	}, nil); !errors.Is(err, sdk.ErrNotFound) {
		return diag.FromErr(err)
	}

	return nil
}

func flattenGroupsWithRoleIDs(sgws *[]sdk.SamlGroupRead) []interface{} {
	if sgws == nil {
		return nil
	}

	groupsWithRoleIDs := make([]interface{}, len(*sgws))

	for i, sgw := range *sgws {
		groupsWithRoleIDs[i] = map[string]interface{}{
			"id":                sgw.Id,
			"looker_group_id":   sgw.LookerGroupId,
			"looker_group_name": sgw.LookerGroupName,
			"name":              sgw.Name,
			"role_ids":          flattenRoles(sgw.Roles),
		}
	}

	return groupsWithRoleIDs
}

func flattenUserAttributesWithIDs(suaws *[]sdk.SamlUserAttributeRead) []interface{} {
	if suaws == nil {
		return nil
	}

	userAttributeWithIDs := make([]interface{}, len(*suaws))

	for i, suaw := range *suaws {
		userAttributeWithIDs[i] = map[string]interface{}{
			"name":               suaw.Name,
			"required":           suaw.Required,
			"user_attribute_ids": flattenUserAttributes(suaw.UserAttributes),
		}
	}

	return userAttributeWithIDs
}

func flattenGroups(groups *[]sdk.Group) []string {
	if groups == nil {
		return []string{}
	}

	groupIDs := make([]string, len(*groups))
	for i, group := range *groups {
		groupIDs[i] = *group.Id
	}

	return groupIDs
}

func flattenRoles(roles *[]sdk.Role) []string {
	if roles == nil {
		return []string{}
	}

	roleIDs := make([]string, len(*roles))
	for i, role := range *roles {
		roleIDs[i] = *role.Id
	}

	return roleIDs
}

func flattenUserAttributes(userAttributes *[]sdk.UserAttribute) []string {
	if userAttributes == nil {
		return []string{}
	}

	userAttributeIDs := make([]string, len(*userAttributes))
	for i, userAttribute := range *userAttributes {
		userAttributeIDs[i] = *userAttribute.Id
	}

	return userAttributeIDs
}
