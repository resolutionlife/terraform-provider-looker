resource "looker_model_set" "test" {
  name   = "test_model_set"
  models = ["test_dataset_1", "test_both_datasets"]
}
