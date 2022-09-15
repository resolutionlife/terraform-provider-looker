# A `looker_role_groups` binding resource can be imported by delimiting the `role_id` and `group_ids` with an underscore. E.g `{{user_id}}_{{group_id}}_{group_id}...}`. 
# See the below syntax. 
terraform import looker_role_groups.writer {{writer_role_id}}_{{writer_group_id}}_{{director_group_id}}
