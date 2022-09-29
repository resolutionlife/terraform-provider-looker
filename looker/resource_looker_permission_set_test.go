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
	// Add a sweeper to remove permission sets that have names starting with `test-acc`.
	resource.AddTestSweepers("looker_permission_set", &resource.Sweeper{
		Name: "looker_permission_set",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
			}

			permissionSets, err := c.SearchPermissionSets(sdk.RequestSearchPermissionSets{
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

			return nil
		},
	})
}

func TestAccLookerPermissionSet(t *testing.T) {
	stop := NewTestProvider("../fixture/looker_permission_set")
	defer stop() //nolint:errcheck

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_permission_set" "test_acc" {
					name        = "test-acc-permission-set"
					permissions = ["access_data", "see_lookml", "see_lookml_dashboards"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_permission_set.test_acc", "name", "test-acc-permission-set"),
					resource.TestCheckResourceAttr("looker_permission_set.test_acc", "permissions.#", "3"),
					testAccPermissionSet("looker_permission_set.test_acc", []string{"access_data", "see_lookml", "see_lookml_dashboards"}),
				),
			},
			{
				Config: `
				resource "looker_permission_set" "test_acc" {
					name        = "test-acc-permission-set"
					permissions = ["see_lookml", "see_lookml_dashboards"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_permission_set.test_acc", "name", "test-acc-permission-set"),
					resource.TestCheckResourceAttr("looker_permission_set.test_acc", "permissions.#", "2"),
					testAccPermissionSet("looker_permission_set.test_acc", []string{"see_lookml", "see_lookml_dashboards"}),
				),
			},
		},
	})
}

func testAccPermissionSet(permSetResource string, expectedPermSets []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permSetRes, ok := s.RootModule().Resources[permSetResource]
		if !ok {
			return errors.Errorf("Not found: %s", permSetResource)
		}
		if permSetRes.Primary.ID == "" {
			return errors.New("permission set ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		permSet, err := client.PermissionSet(permSetRes.Primary.ID, "", nil)
		if err != nil {
			return errors.Wrapf(err, "failed to retrieve permission set with id: %v", permSetRes.Primary.ID)
		}

		if !slice.UnorderedEqual(*permSet.Permissions, expectedPermSets) {
			return errors.Errorf("permissions in permission set do not match expected: %v actual: %v", expectedPermSets, *permSet.Permissions)
		}

		return nil
	}
}
