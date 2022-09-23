package looker

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove user attributes and users that have names starting with `test_acc`.
	resource.AddTestSweepers("looker_user_attribute_user", &resource.Sweeper{
		Name: "looker_user_attribute_user",
		F: func(r string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			userAttrs, err := c.AllUserAttributes(sdk.RequestAllBoardSections{
				Fields: conv.P(""),
				Sorts:  conv.P(""),
			}, nil)
			if err != nil {
				return err
			}

			filteredUserAttrs := make([]sdk.UserAttribute, 0, len(userAttrs))
			for _, userAttr := range userAttrs {
				if strings.HasPrefix(userAttr.Name, "test_acc_") {
					filteredUserAttrs = append(filteredUserAttrs, userAttr)
				}
			}

			for _, filteredUserAttr := range filteredUserAttrs {
				if _, err := c.DeleteUserAttribute(*filteredUserAttr.Id, nil); err != nil {
					return err
				}
			}

			users, err := c.SearchUsers(sdk.RequestSearchUsers{
				Email: conv.P("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, u := range users {
				if _, err := c.DeleteUser(*u.Id, nil); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func TestAccLookerUserAttributeUser(t *testing.T) {
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
					hidden           = true
					default_value    = 24
					user_access      = "View"
					domain_whitelist = ["my_domain/route/sub/*"]
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
					resource.TestCheckResourceAttr("looker_user_attribute_user.test_acc", "value", "25")),
			},
		},
	})
}
