data "looker_model_set" "all" {
  name = "All"
}

data "looker_model_set" "test" {
  id = data.terraform_remote_state.test.outputs.model_set_id
}
