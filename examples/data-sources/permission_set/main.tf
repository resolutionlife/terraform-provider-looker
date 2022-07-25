data "looker_permission_set" "developer" {
  name = "Developer"
}

data "looker_permission_set" "test" {
  id = data.terraform_remote_state.test.outputs.permission_set_id
}
