data "looker_role" "viewer" {
  name = "Viewer"
}

# TODO: data source not yet supported
data "looker_group" "all_users" {
  name = "All Users"
}

resource "looker_saml_config" "saml" {
  enabled    = true
  idp_cert   = "mycert"
  idp_url    = "https://mydomain.com/samlp/metadata/123456"
  idp_issuer = "urn:mydomain.com"

  user_attribute_map_email      = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
  user_attribute_map_first_name = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
  user_attribute_map_last_name  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
  new_user_migration_types      = ["email"]

  default_new_user_role_ids  = [data.looker_role.viewer.id]
  default_new_user_group_ids = [data.looker_group.all_users.id]
}
