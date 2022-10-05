package looker

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
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
	stop := NewTestProvider("../fixture/looker_group_user")
	defer stop() //nolint:errcheck

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

				resource "looker_group" "test_acc" {
					name = "test-acc-group"
				}

				resource "looker_group_user" "test_acc" {
					group_id = looker_group.test_acc.id
					user_id  = looker_user.test_acc.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("looker_group_user.test_acc", "group_id"),
					resource.TestCheckResourceAttrSet("looker_group_user.test_acc", "user_id"),
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
			return fmt.Errorf("Not found: %s", userResource)
		}
		if userRes.Primary.ID == "" {
			return errors.New("user ID is not set")
		}

		groupRes, ok := s.RootModule().Resources[groupResource]
		if !ok {
			return fmt.Errorf("Not found: %s", groupResource)
		}
		if groupRes.Primary.ID == "" {
			return errors.New("group ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		user, err := client.User(userRes.Primary.ID, "", nil)
		if err != nil {
			return fmt.Errorf("failed to retrieve user with id %v: %w", userRes.Primary.ID, err)
		}

		group, err := client.Group(groupRes.Primary.ID, "", nil)
		if err != nil {
			return fmt.Errorf("failed to retrieve group with id %v: %w", userRes.Primary.ID, err)
		}

		if !slice.Contains(*user.GroupIds, *group.Id) {
			return fmt.Errorf("group with ID %v does not contain user %v", *group.Id, *user.Id)
		}

		return nil
	}
}
