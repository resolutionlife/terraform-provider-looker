data "looker_role" "developer" {
  name = "Developer"
}

output "developer" {
  value = data.looker_role.developer.model_set[*].models
}
