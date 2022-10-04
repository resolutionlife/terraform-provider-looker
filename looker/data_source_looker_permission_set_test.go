package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceLookerPermissionSetWithName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_permission_set",
				Config: `
				data "looker_permission_set" "admin" {
					name = "Admin"
				}

				output "looker_permission_set_id" {
					value = data.looker_permission_set.admin.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_permission_set_id", "1"),
				),
			},
		},
	})
}
func TestAccDatasourceLookerPermissionSetWithId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_permission_set",
				Config: `
				data "looker_permission_set" "admin" {
					id = "1"
				}

				output "looker_permission_set_name" {
					value = data.looker_permission_set.admin.name
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_permission_set_name", "Admin"),
				),
			},
		},
	})
}
