resource "looker_user_attribute" "secret_id" {
  name             = "id"
  label            = "secret_id"
  data_type        = "number"
  hidden           = true
  default_value    = 24
  user_access      = "View"
  domain_whitelist = ["my_domain/route/sub/*"]
}
