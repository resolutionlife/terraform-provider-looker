package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLookerUserAPIClient(t *testing.T) {
	stop := NewTestProvider("../fixture/looker_user_api_client")
	defer stop() //nolint:errcheck

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_user" "test_acc" {
				    email      = "test-acc@email.com"
				    first_name = "Tina"
				    last_name  = "Fey"
				  }

				  resource "looker_user_api_client" "test_acc" {
				    user_id = looker_user.test_acc.id
				  }
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("looker_user_api_client.test_acc", "client_id"),
					resource.TestCheckResourceAttrSet("looker_user_api_client.test_acc", "client_secret"),
				),
			},
		},
	})
}
