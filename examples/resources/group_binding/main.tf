 resource "looker_group" "crew" {
  name = "Crew"
}

resource "looker_group" "writers" {
  name = "Writers"
}

resource "looker_group_binding" "crew_writer" {
  parent_group_id = looker_group.crew.id
  group_id        = looker_group.writers.id
}
