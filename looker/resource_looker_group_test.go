package looker

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
)

func init() {
	// Add a sweeper to remove groups that have names starting with `test-acc`.
	resource.AddTestSweepers(
		"looker_group",
		&resource.Sweeper{
			Name: "looker_group",
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
		},
	)
}

func TestAccLookerGroup(t *testing.T) {
	stop := NewTestProvider("../fixture/looker_group")
	defer stop() //nolint:errcheck

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_group" "test_acc" {
					name = "test-acc-group"
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_group.test_acc", "name", "test-acc-group"),
				),
			},
		},
	})
}
