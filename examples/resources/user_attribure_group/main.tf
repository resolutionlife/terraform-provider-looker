resource "looker_group" "test_group_1" {
  name = "Test_1"
}

resource "looker_group" "test_group_2" {
  name = "Test_2"
}

resource "looker_user_attribute" "test_attr" {
  name          = "test_name"
  label         = "test_label"
  data_type     = "number"
  hidden        = false
  default_value = 123
  user_access   = "View"
}

resource "looker_user_attribute_group" "test_group_1" {
  user_attribute_id = looker_user_attribute.test_attr.id
  group_values {
    group_id = looker_group.test_group_1.id
    value    = "0"
  }
  group_values {
    group_id = looker_group.test_group_2.id
    value    = "1"
  }
}
