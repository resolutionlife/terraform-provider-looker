package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceLookerModelSetWithName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_model_set",
				Config: `
				data "looker_model_set" "all" {
					name = "All"
				}

				output "looker_model_set_id" {
					value = data.looker_model_set.all.id
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_model_set_id", "1"),
				),
			},
		},
	})
}

func TestAccDatasourceLookerModelSetWithId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ResourceName: "looker_model_set",
				Config: `
				data "looker_model_set" "all" {
					id = "1"
				}

				output "looker_model_set_name" {
					value = data.looker_model_set.all.name
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("looker_model_set_name", "All"),
				),
			},
		},
	})
}
