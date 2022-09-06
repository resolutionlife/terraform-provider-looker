resource "looker_user" "tina" {
  first_name = "tina"
  last_name  = "fey"
  email      = "tina@orange.com"
}

resource "looker_user_attribute" "secret_id" {
  name             = "id"
  label            = "secret_id"
  data_type        = "number"
  hidden           = true
  default_value    = "24"
  user_access      = "View"
  domain_whitelist = ["my_domain/route/sub/*"]
}

resource "looker_user_attribute_user" "tina_secret_id" {
  user_attribute_id = looker_user_attribute.secret_id.id
  user_id           = looker_user.tina.id
  value             = "23"
}
