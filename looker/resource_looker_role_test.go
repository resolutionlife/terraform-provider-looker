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
	// Add a sweeper to remove model sets, permission sets and roles that have names starting with `test-acc`.
	resource.AddTestSweepers("looker_roles", &resource.Sweeper{
		Name: "looker_roles",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			roles, err := c.SearchRoles(sdk.RequestSearchRoles{
				Name: conv.P("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, role := range roles {
				if _, err := c.DeleteRole(*role.Id, nil); err != nil {
					return err
				}
			}

			permissionSets, err := c.SearchPermissionSets(sdk.RequestSearchPermissionSets{
				Name: conv.P("test-acc%"),
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
				Name: conv.P("test-acc%"),
			}, nil)
			if err != nil {
				return err
			}

			for _, modelSet := range modelSets {
				if _, err := c.DeleteModelSet(*modelSet.Id, nil); err != nil {
					return err
				}
			}

			return nil
		},
	})
}

func TestAccLookerRole(t *testing.T) {
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
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_role.test_acc", "name", "test-acc-role"),
					testAccRole("looker_role.test_acc", []string{"test_dataset_1", "test_dataset_2", "test_both_datasets"}, []string{"access_data", "see_lookml", "see_lookml_dashboards"}),
				),
			},
			{
				Config: `
				resource "looker_permission_set" "test_acc" {
					name        = "test-acc-permission-set"
					permissions = ["see_lookml", "see_lookml_dashboards"]
				}

				resource "looker_model_set" "test_acc" {
					name   = "test-acc-model-set"
					models = ["test_dataset_1", "test_both_datasets"]
				}

				resource "looker_role" "test_acc" {
					name              = "test-acc-role"
					model_set_id      = looker_model_set.test_acc.id
					permission_set_id = looker_permission_set.test_acc.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_role.test_acc", "name", "test-acc-role"),
					testAccRole("looker_role.test_acc", []string{"test_dataset_1", "test_both_datasets"}, []string{"see_lookml", "see_lookml_dashboards"}),
				),
			},
		},
	})
}

func testAccRole(roleResource string, expectedModelsSets, expectedPermSets []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		roleRes, ok := s.RootModule().Resources[roleResource]
		if !ok {
			return errors.Errorf("Not found: %s", roleResource)
		}
		if roleRes.Primary.ID == "" {
			return errors.New("role ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		role, err := client.Role(roleRes.Primary.ID, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve role with id: %v", roleRes.Primary.ID)
		}

		if !slice.UnorderedEqual(*role.PermissionSet.Permissions, expectedPermSets) {
			return errors.Errorf("permissions in role do not match expected permissions: %v actual: %v", expectedPermSets, *role.PermissionSet.Permissions)
		}

		if !slice.UnorderedEqual(*role.ModelSet.Models, expectedModelsSets) {
			return errors.Errorf("models in role do not match expected models: %v actual: %v", expectedModelsSets, *role.ModelSet.Models)
		}

		return nil
	}
}
