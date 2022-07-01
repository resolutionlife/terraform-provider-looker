provider "looker" {}

resource "looker_user" "my-user" {
  user_email = "test@example.com"
  first_name = "FirstName"
  last_name  = "LastName"
}
