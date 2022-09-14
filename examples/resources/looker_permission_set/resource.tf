
resource "looker_permission_set" "developer" {
  name        = "Developer"
  permissions = ["access_data", "see_lookml", "see_lookml_dashboards"]
}
