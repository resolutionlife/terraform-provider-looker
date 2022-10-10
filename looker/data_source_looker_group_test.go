package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceLookerGroup(t *testing.T) {
	stop := NewTestProvider("../fixture/looker_data_group")
	//nolint:errcheck
	defer stop()

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_group",
				Config: `
				data "looker_group" "all" {
					name = "All Users"
				}

				output "looker_group_all_id" {
					value = data.looker_group.all.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_group_all_id", "1"),
				),
			},
			{
				ResourceName: "looker_group",
				Config: `
				data "looker_group" "all_id" {
					id = "1"
				}

				output "looker_group_all_name" {
					value = data.looker_group.all_id.name
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_group_all_name", "All Users"),
				),
			},
		},
	})
}
