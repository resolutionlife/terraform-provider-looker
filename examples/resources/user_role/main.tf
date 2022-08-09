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

resource "looker_user_role" "tina_roles" {
  user_id  = looker_user.tina.id
  role_ids = [looker_role.director.id, looker_role.producer.id]
}
