resource "looker_group" "directors" {
  name = "Directors"
}

resource "looker_group" "writers" {
  name = "Writers"
}

resource "looker_group_binding" "director_writer" {
  parent_group_id = looker_group.directors.id
  group_id        = looker_group.writers.id
}
