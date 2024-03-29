---
page_title: "looker_user_roles Resource - terraform-provider-looker"
subcategory: ""
description: |-
  This resource binds a set of roles to a Looker user. This is an additive and non-authorative resource that grants roles in addition to current roles configured in Looker.
---

# looker_user_roles (Resource)

This resource binds a set of roles to a looker user. This is an **additive and non-authorative** resource that grants roles **in addition** to current roles configured in Looker.

~>The `looker_user_roles` resource **cannot** be used in conjunction with another `looker_user_roles` resource if they grant privileges to the same user, otherwise they will fight over what roles should be set.

## Example Usage

```terraform
resource "looker_user" "tina" {
  first_name = "tina"
  last_name  = "fey"
  email      = "tina@orange.com"
}

data "looker_role" "director" {
  name = "Director"
}

data "looker_role" "producer" {
  name = "Producer"
}

resource "looker_user_roles" "tina_roles" {
  user_id  = looker_user.tina.id
  role_ids = [looker_role.director.id, looker_role.producer.id]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `role_ids` (Set of String) A slice of role_ids which will be assigned to the user
- `user_id` (String) The id of the user

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# A `looker_user_roles` binding resource can be imported using the following syntax:

terraform import looker_user_roles.tina_roles {{user_id}}
```
