data "looker_role" "writer" {
  name = "Writer"
}

resource "looker_group" "writers" {
  name = "Writers"
}

resource "looker_group" "directors" {
  name = "Directors"
}

# give writers and directors the writer role
resource "looker_role_groups" "tina-director" {
  role_id   = looker_role.writer.id
  group_ids = [looker_group.writer.id, looker_group.directors.id]
}
