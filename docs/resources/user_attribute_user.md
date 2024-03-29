---
page_title: "looker_user_attribute_user Resource - terraform-provider-looker"
subcategory: ""
description: |-
  This resource sets a value onto a user for the given user attribute. If a default value is already set for the user attribute, this value will override the default value. Note that if the user attribute values are hidden (can be configured when provisioning a looker_user_attribute) then the provider does not have the permissions to read the hidden values, and cannot verify if the value has been manually changed in the Looker UI. The provider can however check if the value has been removed, and will prompt to recreate the resource.
---

# looker_user_attribute_user (Resource)

This resource sets a value onto a user for the given user attribute. If a default value is already set for the user attribute, this value will override the default value. 

~>If the user attribute values are hidden (can be configured when provisioning a `looker_user_attribute`) then the provider does not have the permissions to read the hidden values, and cannot verify if the value has been manually changed in the Looker UI. The provider can however check if the value has been removed, and will prompt to recreate the resource.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `user_attribute_id` (String) The id of the user attribute
- `user_id` (String) The id of the user
- `value` (String) The value of the user attribute to be set on the user

### Read-Only

- `id` (String) The id of the resource. This id is of the form <user_attribute_id>_<user_id>

## Import

Import is supported using the following syntax:

~> Imports are not supported for a `user_attribute` with `hidden = true` as the API does not have the permissions to read the hidden values. One method to import would be to reapply the changes after the import is successful. 

```shell
# A `looker_user_attribute_user` resource can be imported by delimiting the `user_attribute_id` and `user_id` with an underscore. E.g `{{user_attribute_id}}_{{user_id}}`. 
# See the below syntax. 
terraform import looker_user_attribute_user.tina_employee_number {{user_attribute_id}}_{{user_id}}
```
