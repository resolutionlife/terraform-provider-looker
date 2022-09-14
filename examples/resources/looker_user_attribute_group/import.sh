# A `looker_user_attribute_group` resource can be imported by delimiting the `user_attribute_id` with a list of `group_id` with an underscore. E.g `{{user_attribute_id}}_{{group_id}}_{{...}}`. 
# See the below syntax.
terraform import looker_user_attribute_group.test_group {{user_attribute_id}}_{{group_id}}