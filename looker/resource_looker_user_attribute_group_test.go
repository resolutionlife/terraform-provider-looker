package looker

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
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
					hidden           = false
					default_value    = 24
					user_access      = "View"
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
					testAccUserAttributeGroup("looker_user_attribute_group.test_acc", "looker_group.test_acc", "25"),
				),
			},
		},
	})
}

func testAccUserAttributeGroup(userAttrGroupResource, groupResource, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userAttrGroupRes, ok := s.RootModule().Resources[userAttrGroupResource]
		if !ok {
			return fmt.Errorf("Not found: %s", userAttrGroupResource)
		}
		if userAttrGroupRes.Primary.ID == "" {
			return errors.New("user attribute group ID is not set")
		}

		groupRes, ok := s.RootModule().Resources[groupResource]
		if !ok {
			return fmt.Errorf("Not found: %s", groupRes)
		}
		if groupRes.Primary.ID == "" {
			return errors.New("group ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		userAttrGroupIds := strings.Split(userAttrGroupRes.Primary.ID, "_")
		if len(userAttrGroupIds) < 2 {
			return errors.New("invalid id, should be of the form <user_attribute_id>_<group_id>_<...>")
		}

		// id of user attribute is in form <user_attribute_id>_<group_id>_<...>
		userAttrs, err := client.AllUserAttributeGroupValues(userAttrGroupIds[0], "", nil)
		if err != nil {
			return fmt.Errorf("failed to retrieve user attribute group value with id %v: %w", userAttrGroupRes.Primary.ID, err)
		}

		for _, userAttr := range userAttrs {
			if *userAttr.GroupId == groupRes.Primary.ID {
				if *userAttr.Value != expectedValue {
					return fmt.Errorf("value in user attribute group does not match expected: %v actual: %v", expectedValue, *userAttr.Value)
				}
			}
		}

		return nil
	}
}
