
resource "looker_permission_set" "writer" {
  name        = "Writer"
  permissions = ["access_data", "see_lookml", "see_lookml_dashboards"]
}
