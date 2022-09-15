resource "looker_user" "user" {
  first_name = "Tina"
  last_name  = "Fey"
  email      = "tina@orange.com"
}

resource "looker_user_api_client" "client" {
  user_id = looker_user.user.id
}
