resource "looker_user_attribute" "secret_id" {
  name             = "id"
  label            = "Secret ID"
  data_type        = "number"
  hidden           = true
  default_value    = 24
  user_access      = "View"
  domain_whitelist = ["my_domain/route/sub/*"]
}

resource "looker_user_attribute" "employee_number" {
  name          = "employee_number"
  label         = "Employee Number"
  data_type     = "number"
  hidden        = false
  default_value = "24"
  user_access   = "View"
}
