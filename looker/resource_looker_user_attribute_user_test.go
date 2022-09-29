package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/looker-open-source/sdk-codegen/go/rtl"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove user attributes and users that have names starting with `test_acc`.
	resource.AddTestSweepers("looker_user_attribute_user", &resource.Sweeper{
		Dependencies: []string{"looker_user_attribute", "looker_user"},
		Name:         "looker_user_attribute_user",
		F: func(r string) error {
			return nil
		},
	})
}

func TestAccLookerUserAttributeUser(t *testing.T) {
	stop := NewTestProvider("../fixture/looker_user_attribute_user")
	defer stop()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_user" "test_acc" {
				    email      = "test-acc@email.com"
				    first_name = "John"
				    last_name  = "Doe"
				}

				resource "looker_user_attribute" "test_acc" {
					name             = "test_acc_user_attribute_name"
					label            = "test-acc-user-attribute-label"
					data_type        = "number"
					hidden           = false
					default_value    = 24
					user_access      = "View"
				}

				resource "looker_user_attribute_user" "test_acc" {
					user_attribute_id = looker_user_attribute.test_acc.id
					user_id           = looker_user.test_acc.id
					value             = 25
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("looker_user_attribute_user.test_acc", "user_attribute_id"),
					resource.TestCheckResourceAttrSet("looker_user_attribute_user.test_acc", "user_id"),
					resource.TestCheckResourceAttr("looker_user_attribute_user.test_acc", "value", "25"),
					testAccRoleUserAttributeUser("looker_user.test_acc", "looker_user_attribute_user.test_acc", "25"),
				),
			},
		},
	})
}

func testAccRoleUserAttributeUser(userResource, userAttrUserResource, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userRes, ok := s.RootModule().Resources[userResource]
		if !ok {
			return errors.Errorf("Not found: %s", userResource)
		}
		if userRes.Primary.ID == "" {
			return errors.New("user ID is not set")
		}

		userAttrUserRes, ok := s.RootModule().Resources[userAttrUserResource]
		if !ok {
			return errors.Errorf("Not found: %s", userAttrUserResource)
		}
		if userAttrUserRes.Primary.ID == "" {
			return errors.New("user attribute ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		// UserAttributeIds is broken, cannot filter for a specific user attribute
		userAttrs, err := client.UserAttributeUserValues(sdk.RequestUserAttributeUserValues{
			UserId:           userRes.Primary.ID,
			UserAttributeIds: &rtl.DelimString{userAttrUserRes.Primary.Attributes["user_attribute_id"]},
			Fields:           conv.P(""),
		}, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve user attribute values with id: %v", userRes.Primary.ID)
		}

		for _, userAttr := range userAttrs {
			if *userAttr.UserAttributeId == userAttrUserRes.Primary.Attributes["user_attribute_id"] {
				if *userAttr.Value != expectedValue {
					return errors.Errorf("value in user attribute does not match expected: %v actual: %v", expectedValue, *userAttr.Value)
				}
			}
		}

		return nil
	}
}
