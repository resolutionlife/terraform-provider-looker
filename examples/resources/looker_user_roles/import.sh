# A `looker_user_roles` binding resource can be imported by delimiting the `user_id` and `role_ids` with an underscore. E.g `{{user_id}}_{{role_ids}}`. 
# See the below syntax. 
terraform import looker_user_roles.tina_roles {{user_id}}_{{role_ids}}
