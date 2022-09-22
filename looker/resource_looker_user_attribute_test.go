package looker

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove user attributes that have names starting with `test_acc`.
	resource.AddTestSweepers("looker_user_attribute", &resource.Sweeper{
		Name: "looker_user_attribute",
		F: func(r string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			userAttrs, err := c.AllUserAttributes(sdk.RequestAllBoardSections{
				Fields: conv.PString(""),
				Sorts:  conv.PString(""),
			}, nil)
			if err != nil {
				return err
			}

			filteredUserAttrs := make([]sdk.UserAttribute, 0)
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

			return nil
		},
	})
}

func TestAccLookerUserAttribute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_user_attribute" "test_acc" {
					name             = "test_acc_user_attribute_name"
					label            = "test-acc-user-attribute-label"
					data_type        = "number"
					hidden           = true
					default_value    = 24
					user_access      = "View"
					domain_whitelist = ["my_domain/route/sub/*"]
				}  
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "name", "test_acc_user_attribute_name"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "label", "test-acc-user-attribute-label"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "data_type", "number"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "hidden", "true"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "default_value", "24"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "user_access", "View"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "domain_whitelist.#", "1"),
				),
			},
			// only update values that do NOT trigger a recreate
			{
				Config: `
				resource "looker_user_attribute" "test_acc" {
					name             = "test_acc_user_attribute_name"
					label            = "test-acc-user-attribute-label"
					data_type        = "string"
					hidden           = true
					default_value    = "abc"
					user_access      = "View"
					domain_whitelist = ["my_domain/route/sub/*"]
				}  
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "name", "test_acc_user_attribute_name"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "label", "test-acc-user-attribute-label"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "data_type", "string"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "hidden", "true"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "default_value", "abc"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "user_access", "View"),
					resource.TestCheckResourceAttr("looker_user_attribute.test_acc", "domain_whitelist.#", "1"),
				),
			},
		},
	})
}
