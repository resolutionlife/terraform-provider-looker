A `looker_group_user` binding resource can be imported by delimiting the `user_id` and `group_id` with an underscore. E.g `{{user_id}}_{{group_id}}`. See the below syntax. 

```
terraform import looker_group_user.tina_director {{user_id}}_{{group_id}}