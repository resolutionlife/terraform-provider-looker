resource "looker_permission_set" "writer" {
  name        = "Writer"
  permissions = ["see_lookml", "see_lookml_dashboards"]
}

resource "looker_model_set" "writer" {
  name   = "Writer"
  models = ["test_dataset_1", "test_both_datasets"]
}

resource "looker_roles" "writer" {
  name              = "Writer"
  model_set_id      = looker_model_set.test.id
  permission_set_id = looker_permission_set.test.id
}
