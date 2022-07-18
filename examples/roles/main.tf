# TODO: extend this example to build the looker_model_set and looker_permission_set resources when they are supported

provider "looker" {}

resource "looker_roles" "test" {
  name              = "Test Role"
  model_set_id      = looker_model_set.test.id
  permission_set_id = looker_permission_set.test.id
}
