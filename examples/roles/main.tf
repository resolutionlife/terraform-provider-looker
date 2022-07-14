# TODO: extend this example to build the test_model_set and test_permission_set resources when they are supported

resource "looker_roles" "test" {
  name              = "Test Role"
  model_set_id      = resource.test_model_set.id
  permission_set_0d = resource.test_permission_set.id
}
