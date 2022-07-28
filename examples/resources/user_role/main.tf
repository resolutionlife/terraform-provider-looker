
resource "looker_user" "tina" {
  first_name = "tina"
  last_name  = "fey"
  email      = "tina@orange.com"
}

data "looker_permission_set" "viewer" {
  name = "Viewer"
}

data "looker_model_set" "director" {
  name = "director"
}

resource "looker_role" "director" {
  name              = "Director"
  permission_set_id = data.looker_permission_set.viewer.id
  model_set_id      = data.looker_model_set.director.id
}

resource "looker_user_role" "test-developer" {
  user_id  = looker_user.tina.id
  role_ids = [looker_role.director.id]
}
