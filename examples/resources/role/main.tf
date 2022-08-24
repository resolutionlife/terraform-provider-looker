resource "looker_permission_set" "test" {
  name        = "test_permission_set"
  permissions = ["see_lookml", "see_lookml_dashboards"]
}

resource "looker_model_set" "test" {
  name   = "test_model_set"
  models = ["test_dataset_1", "test_both_datasets"]
}

resource "looker_roles" "test" {
  name              = "Test Role"
  model_set_id      = looker_model_set.test.id
  permission_set_id = looker_permission_set.test.id
}
