data "looker_role" "developer" {
  name = "Developer"
}

output "developer_model_set_models" {
  value = data.looker_role.developer.model_set[*].models
}
