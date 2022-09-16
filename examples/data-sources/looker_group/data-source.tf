data "looker_group" "all" {
  name = "All Users"
}

output "all_users_id" {
  value = data.looker_group.all.id
}

data "looker_group" "all_id" {
  id = "1"
}

output "all_users_name" {
  value = data.looker_group.all_id.name
}
