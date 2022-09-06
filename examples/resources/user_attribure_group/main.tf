resource "looker_group" "test_group" {
  name = "Test"
}

resource "looker_user_attribute" "test_attr" {
  name          = "test_name"
  label         = "test_label"
  data_type     = "number"
  hidden        = false
  default_value = 123
  user_access   = "View"
}

resource "looker_user_attribute_group" "test_attr_group" {
  group_id          = looker_group.test_group.id
  user_attribute_id = looker_user_attribute.test_attr.id
  value             = "2"
}
