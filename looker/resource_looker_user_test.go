package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v4 "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove users that have emails starting with `test-acc`.
	resource.AddTestSweepers("looker_user", &resource.Sweeper{
		Name: "looker_user",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			users, err := c.SearchUsers(v4.RequestSearchUsers{
				Email: conv.PString("test-acc%"),
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

func TestAccLookerUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				  resource "looker_user" "test_acc" {
				    email      = "test-acc@resolutionlife.com"
				    first_name = "John"
				    last_name  = "Doe"
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_user.test_acc", "email", "test-acc@resolutionlife.com"),
					resource.TestCheckResourceAttr("looker_user.test_acc", "first_name", "John"),
					resource.TestCheckResourceAttr("looker_user.test_acc", "last_name", "Doe"),
				),
			},
			{
				Config: `
				  resource "looker_user" "test_acc" {
				    email      = "test-acc@resolutionlife.com"
				    first_name = "Jane"
				    last_name  = "Smith"
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_user.test_acc", "first_name", "Jane"),
					resource.TestCheckResourceAttr("looker_user.test_acc", "last_name", "Smith"),
				),
			},
		},
	})
}
