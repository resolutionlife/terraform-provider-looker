data "looker_idp_metadata" "md" {
  idp_metadata_url = "https://auth.com/samlp/metadata/123456"
}

data "looker_idp_metadata" "md" {
  idp_metadata_xml = "<?xml ...>"
}
