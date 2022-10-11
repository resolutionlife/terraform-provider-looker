resource "looker_model_set" "writer" {
  name   = "Writer"
  models = ["test_dataset_1", "test_both_datasets"]
}
