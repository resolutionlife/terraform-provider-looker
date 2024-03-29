package looker

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"

	"github.com/resolutionlife/terraform-provider-looker/internal/conv"
	"github.com/resolutionlife/terraform-provider-looker/internal/slice"
)

func init() {
	// Add a sweeper to remove model sets that have names starting with `test-acc`.
	resource.AddTestSweepers("looker_model_set", &resource.Sweeper{
		Name: "looker_model_set",
		F: func(_ string) error {
			c, err := newTestLookerSDK()
			if err != nil {
				return err
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

			return nil
		},
	})
}

func TestAccLookerModelSet(t *testing.T) {
	stop := NewTestProvider("../fixture/looker_model_set")
	defer stop() //nolint:errcheck

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "looker_model_set" "test_acc" {
					name   = "test-acc-model-set"
					models = ["test_dataset_1", "test_dataset_2", "test_both_datasets"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_model_set.test_acc", "name", "test-acc-model-set"),
					resource.TestCheckResourceAttr("looker_model_set.test_acc", "models.#", "3"),
					testAccModelSet("looker_model_set.test_acc", []string{"test_dataset_1", "test_dataset_2", "test_both_datasets"}),
				),
			},
			{
				Config: `
				resource "looker_model_set" "test_acc" {
					name   = "test-acc-model-set"
					models = ["test_dataset_1", "test_both_datasets"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("looker_model_set.test_acc", "name", "test-acc-model-set"),
					resource.TestCheckResourceAttr("looker_model_set.test_acc", "models.#", "2"),
					testAccModelSet("looker_model_set.test_acc", []string{"test_dataset_1", "test_both_datasets"}),
				),
			},
		},
	})
}

func testAccModelSet(modelSetResource string, expectedModelsSets []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		modelSetRes, ok := s.RootModule().Resources[modelSetResource]
		if !ok {
			return fmt.Errorf("Not found: %s", modelSetResource)
		}
		if modelSetRes.Primary.ID == "" {
			return errors.New("model set ID is not set")
		}

		client := testAccProvider.Meta().(*sdk.LookerSDK)

		modelSet, err := client.ModelSet(modelSetRes.Primary.ID, "", nil)
		if err != nil {
			return fmt.Errorf("failed to retrieve model set with id %v: %w", modelSetRes.Primary.ID, err)
		}

		if !slice.UnorderedEqual(*modelSet.Models, expectedModelsSets) {
			return fmt.Errorf("models in model set do not match expected: %v actual: %v", expectedModelsSets, *modelSet.Models)
		}

		return nil
	}
}
