# A `looker_group_group` resource can be imported by delimiting the `parent_group_id` and `group_id` with an underscore. E.g `{{parent_group_id}}_{{group_id}}`. See the below syntax. 
terraform import looker_group_group.crew_writer {{crew_group_id}}_{{writer_group_id}}
