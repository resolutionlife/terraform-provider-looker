package looker

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/pkg/errors"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove roles, groups and role groups that have names starting with `test-acc`.
	// TODO: Maybe move these sweepers to a dedicated test utils package for re-usability
	resource.AddTestSweepers("looker_role_groups", &resource.Sweeper{
		Name: "looker_role_groups",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			permissionSets, err := c.SearchPermissionSets(sdk.RequestSearchModelSets{
				Name: conv.PString("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, permissionSet := range permissionSets {
				if _, err := c.DeletePermissionSet(*permissionSet.Id, nil); err != nil {
					return err
				}
			}

			modelSets, err := c.SearchModelSets(sdk.RequestSearchModelSets{
				Name: conv.PString("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, modelSet := range modelSets {
				if _, err := c.DeleteModelSet(*modelSet.Id, nil); err != nil {
					return err
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

			roles, err := c.SearchRoles(sdk.RequestSearchRoles{
				Name: conv.PString("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, role := range roles {
				if _, err := c.DeleteRole(*role.Id, nil); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func TestAccLookerRoleGroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_permission_set" "test_acc" {
					name        = "test-acc-permission-set"
					permissions = ["access_data", "see_lookml", "see_lookml_dashboards"]
				}

				resource "looker_model_set" "test_acc" {
					name   = "test-acc-model-set"
					models = ["test_dataset_1", "test_dataset_2", "test_both_datasets"]
				}

				resource "looker_role" "test_acc" {
					name              = "test-acc-role"
					model_set_id      = looker_model_set.test_acc.id
					permission_set_id = looker_permission_set.test_acc.id
				}

				resource "looker_group" "test_acc_1" {
					name = "test-acc-group-1"
				}

				resource "looker_group" "test_acc_2" {
					name = "test-acc-group-2"
				}

				resource "looker_role_groups" "test_acc" {
					role_id   = looker_role.test_acc.id
					group_ids = [looker_group.test_acc_1.id, looker_group.test_acc_2.id]
				} 
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_role_groups.test_acc", "group_ids.#", "2"),
					testAccRoleGroups("looker_role_groups.test_acc"),
				),
			},
		},
	})
}

func testAccRoleGroups(roleGroupResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		roleGroupsRes, ok := s.RootModule().Resources[roleGroupResource]
		if !ok {
			return errors.Errorf("Not found: %s", roleGroupResource)
		}
		if roleGroupsRes.Primary.ID == "" {
			return errors.New("role groups ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		roleGroups, err := client.RoleGroups(roleGroupsRes.Primary.ID, "", nil)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve role group with id: %v", roleGroupsRes.Primary.ID)
		}

		var groupIds []string
		for _, roleGroup := range roleGroups {
			groupIds = append(groupIds, *roleGroup.ExternalGroupId)
		}

		spew.Dump(groupIds)

		return nil
	}
}
