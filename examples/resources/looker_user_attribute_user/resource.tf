resource "looker_user" "tina" {
  first_name = "tina"
  last_name  = "fey"
  email      = "tina@orange.com"
}

resource "looker_user_attribute" "employee_number" {
  name          = "employee_number"
  label         = "Employee Number"
  data_type     = "number"
  hidden        = false
  default_value = "0"
  user_access   = "View"
}

resource "looker_user_attribute_user" "tina_employee_number" {
  user_attribute_id = looker_user_attribute.employee_number.id
  user_id           = looker_user.tina.id
  value             = "23"
}
