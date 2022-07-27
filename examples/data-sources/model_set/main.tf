data "looker_model_set" "all" {
  name = "all"
}

data "looker_model_set" "test" {
  id = data.terraform_remote_state.test.outputs.model_set_id
}
