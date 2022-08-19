resource "looker_group" "directors" {
  name = "Directors"
}

resource "looker_user" "tina" {
  first_name = "tina"
  last_name  = "fey"
  email      = "tina@orange.com"
}

resource "looker_group_user" "tina-director" {
  group_id = looker_group.directors.id
  user_id  = looker_user.tina.id
}
