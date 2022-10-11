package looker

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "This data source reads a Looker role from a Looker instance.",

		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "The name of the role. This field is case sensitive. See the documentation on looker [roles](https://docs.looker.com/admin-options/settings/roles) and [default roles](https://docs.looker.com/admin-options/settings/roles#default_roles) generated by Looker.",
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "The id of the role",
			},
			"permission_set": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the permission set",
						},
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the permission set",
						},
						"permissions": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed:    true,
							Description: "An unordered list of models within the permission set",
						},
					},
				},
				Computed:    true,
				Description: "The permission set binded to the looker role. This `permission_set` attribute holds additional attributes about the permission set.",
			},
			"model_set": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the model set. This field is case sensitive",
						},
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the model set",
						},
						"models": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed:    true,
							Description: "An unordered list of models within the model set.",
						},
					},
				},
				Computed:    true,
				Description: "The model set binded to the looker role. This `model_set` attribute holds additional attributes about the model set",
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// exactly one of these variables will be nil - this is enforced by the data source schema
	name := conv.PString(d.Get("name").(string))
	id := conv.PString(d.Get("id").(string))

	roles, roleErr := api.SearchRoles(sdk.RequestSearchRoles{
		Name: name,
		Id:   id,
	}, nil)
	if roleErr != nil {
		return diag.FromErr(roleErr)
	}

	var role *sdk.Role
	for _, r := range roles {
		if id != nil && r.Id != nil && *id == *r.Id {
			role = &r
			break
		}

		if name != nil && r.Name != nil && *name == *r.Name {
			role = &r
			break
		}
	}
	if role == nil {
		return diag.Errorf("role not found")
	}
	d.SetId(*role.Id)

	result := multierror.Append(
		d.Set("name", role.Name),
		d.Set("permission_set", flattenPermissionSet(role.PermissionSet)),
		d.Set("model_set", flattenModelSet(role.ModelSet)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

// flattenModelSet takes a model set object and maps the field from the object to a map[string]interface{}. The key is the attribute name defined in the data source
// schema, and the value is the corresponding value coming from the looker api. This function returns a []interface{} because the
// terraform aggregate type schema.TypeSet expects a slice of attributes.
func flattenModelSet(m *sdk.ModelSet) []interface{} {
	ms := make(map[string]interface{}, 0)

	ms["name"] = m.Name
	ms["id"] = m.Id

	if m.Models != nil {
		ms["models"] = *m.Models
	}

	return []interface{}{ms}
}

// flattenPermissionSet takes a permission set object and maps the field from the object to a map[string]interface{}. The key is the attribute name defined in the data source
// schema, and the value is the corresponding value coming from the looker api. This function returns a []interface{} because the
// terraform aggregate type schema.TypeSet expects a slice of attributes.
func flattenPermissionSet(p *sdk.PermissionSet) []interface{} {
	ps := make(map[string]interface{}, 0)

	ps["name"] = p.Name
	ps["id"] = p.Id

	if p.Permissions != nil {
		ps["permissions"] = *p.Permissions
	}

	return []interface{}{ps}
}
