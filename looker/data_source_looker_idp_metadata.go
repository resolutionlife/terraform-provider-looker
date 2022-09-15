package looker

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/resolutionlife/terraform-provider-looker/internal/conv"

	sdk "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

func datasourceLookerIdpMetadata() *schema.Resource {
	return &schema.Resource{
		Description: "This datasource parsed IdP metadata.",

		ReadContext: datasourceIdpMetadataRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"idp_metadata_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"idp_metadata_url", "idp_metadata_xml"},
				Description:  "SAML Metadata URL",
			},
			"idp_metadata_xml": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"idp_metadata_url", "idp_metadata_xml"},
				Description:  "SAML Metadata XML",
			},
			"idp_issuer": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identify Provider Issuer",
			},
			"idp_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identify Provider URL",
			},
			"idp_cert": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identify Provider Certificate",
			},
		},
	}
}

func datasourceIdpMetadataRead(ctx context.Context, d *schema.ResourceData, c interface{}) diag.Diagnostics {
	api := c.(*sdk.LookerSDK)

	// exactly one of these variables will be nil - this is enforced by the data source schema
	url := conv.PString(d.Get("idp_metadata_url").(string))
	xml := conv.PString(d.Get("idp_metadata_xml").(string))

	var res sdk.SamlMetadataParseResult
	var err error

	switch {
	case xml != nil:
		res, err = api.ParseSamlIdpMetadata(*xml, nil)
	case url != nil:
		res, err = api.FetchAndParseSamlIdpMetadata(*url, nil)
	default:
		err = errors.New("missing URL or XML")
	}

	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("idp_url", res.IdpUrl),
		d.Set("idp_issuer", res.IdpIssuer),
		d.Set("idp_cert", res.IdpCert),
	)

	d.SetId(*res.IdpUrl)
	return diag.FromErr(result.ErrorOrNil())
}
