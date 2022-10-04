package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: Check output of model_sets and permission_sets tied to the looker_role
func TestAccDatasourceLookerRoleWithId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_role",
				Config: `
				data "looker_role" "admin" {
					name = "Admin"
				}

				output "looker_role_id" {
					value = data.looker_role.admin.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_role_id", "2"),
				),
			},
		},
	})
}

func TestAccDatasourceLookerRoleWithName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_role",
				Config: `
				data "looker_role" "admin" {
					id = "2"
				}

				output "looker_role_name" {
					value = data.looker_role.admin.name
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_role_name", "Admin"),
				),
			},
		},
	})
}
