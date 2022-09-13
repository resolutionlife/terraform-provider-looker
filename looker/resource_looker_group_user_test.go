package looker

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	client "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove groups and users that have names starting with `test-acc`.
	resource.AddTestSweepers("looker_group_user", &resource.Sweeper{
		Name: "looker_group_user",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			users, err := c.SearchUsers(sdk.RequestSearchUsers{
				Email: conv.PString("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, u := range users {
				if _, err := c.DeleteUser(*u.Id, nil); err != nil {
					return err
				}
				for _, groupId := range *u.GroupIds {
					if err := c.DeleteGroupUser(groupId, *u.Id, nil); err != nil {
						return err
					}
				}
			}

			groups, err := c.SearchGroups(sdk.RequestSearchGroups{
				Name: conv.PString("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, g := range groups {
				if _, err := c.DeleteGroup(*g.Id, nil); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func TestAccLookerGroupUser(t *testing.T) {
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

				resource "looker_group" "test_acc" {
					name = "test-acc-group"
				}

				resource "looker_group_user" "test_acc" {
					group_id = looker_group.test_acc.id
					user_id  = looker_user.test_acc.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccGroupUserBinding("looker_user.test_acc", "looker_group.test_acc"),
				),
			},
		},
	})
}

func testAccGroupUserBinding(userResource, groupResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		userRes, ok := s.RootModule().Resources[userResource]
		if !ok {
			return errors.Errorf("Not found: %s", userResource)
		}
		if userRes.Primary.ID == "" {
			return errors.New("user ID is not set")
		}

		groupRes, ok := s.RootModule().Resources[groupResource]
		if !ok {
			return errors.Errorf("Not found: %s", groupResource)
		}
		if groupRes.Primary.ID == "" {
			return errors.New("grou[] ID is not set")
		}

		client := testAccProvider.Meta().(*client.LookerSDK)

		_, err := client.User(userRes.Primary.ID, "", nil)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve user with id: %v", userRes.Primary.ID)
		}

		// TODO: Maybe find group and ensure user is part of said group?

		return nil
	}
}