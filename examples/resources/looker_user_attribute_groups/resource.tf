resource "looker_group" "writers" {
  name = "Writers"
}

resource "looker_group" "interns" {
  name = "Interns"
}

resource "looker_user_attribute" "employee_number" {
  name          = "employee_number"
  label         = "Employee Number"
  data_type     = "number"
  hidden        = false
  default_value = "0"
  user_access   = "View"
}

resource "looker_user_attribute_groups" "employee_number" {
  user_attribute_id = looker_user_attribute.employee_number.id
  group_values {
    group_id = looker_group.writers.id
    value    = "1"
  }
  group_values {
    group_id = looker_group.interns.id
    value    = "2"
  }
}
