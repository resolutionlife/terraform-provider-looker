package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func init() {
	// Add a sweeper to remove groups that have names starting with `test-acc`.
	resource.AddTestSweepers("looker_group_group", &resource.Sweeper{
		Name: "looker_group_group",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
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

func TestAccLookerGroupGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_group" "test_acc_parent" {
					name = "test-acc-parent-group"
				}

				resource "looker_group" "test_acc_child" {
					name = "test-acc-child-group"
				}

				resource "looker_group_group" "test_acc_group" {
					parent_group_id = looker_group.test_acc_parent.id
					group_id  = looker_group.test_acc_child.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("looker_group_group.test_acc_group", "parent_group_id"),
					resource.TestCheckResourceAttrSet("looker_group_group.test_acc_group", "group_id"),
					testAccGroupGroupBinding("looker_group.test_acc_parent", "looker_group.test_acc_child"),
				),
			},
		},
	})
}

func testAccGroupGroupBinding(parentGroupResource, childGroupResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		parentRes, ok := s.RootModule().Resources[parentGroupResource]
		if !ok {
			return errors.Errorf("Not found: %s", parentGroupResource)
		}
		if parentRes.Primary.ID == "" {
			return errors.New("parent group ID is not set")
		}

		childRes, ok := s.RootModule().Resources[childGroupResource]
		if !ok {
			return errors.Errorf("Not found: %s", childGroupResource)
		}
		if childRes.Primary.ID == "" {
			return errors.New("child group ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		parentSubGroups, err := client.AllGroupGroups(parentRes.Primary.ID, "", nil)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve parent group with id: %v", parentRes.Primary.ID)
		}

		groupIds := []string{}
		for _, subGroup := range parentSubGroups {
			groupIds = append(groupIds, *subGroup.Id)
		}

		if !slice.Contains(groupIds, childRes.Primary.ID) {
			return errors.Errorf("parent group contains groups with ID %v does not contain group with ID %v", groupIds, childRes.Primary.ID)
		}

		return nil
	}
}
