package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	// Add a sweeper to remove user attributes and groups that have names starting with `test_acc`.
	resource.AddTestSweepers("looker_user_attribute_group", &resource.Sweeper{
		Dependencies: []string{"looker_user_attribute", "looker_group"},
		Name:         "looker_user_attribute_group",
		F: func(r string) error {
			return nil
		},
	})
}

func TestAccLookerUserAttributeGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_group" "test_acc" {
					name = "test-acc-group"
				}

				resource "looker_user_attribute" "test_acc" {
					name             = "test_acc_user_attribute_name"
					label            = "test-acc-user-attribute-label"
					data_type        = "number"
					hidden           = true
					default_value    = 24
					user_access      = "View"
					domain_whitelist = ["my_domain/route/sub/*"]
				}

				resource "looker_user_attribute_group" "test_acc" {
					user_attribute_id = looker_user_attribute.test_acc.id
					group_values {
					  group_id = looker_group.test_acc.id
					  value    = "25"
					}
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("looker_user_attribute_group.test_acc", "user_attribute_id"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"looker_user_attribute_group.test_acc",
						"group_values.*",
						map[string]string{
							"value": "25",
						},
					),
				),
			},
		},
	})
}
