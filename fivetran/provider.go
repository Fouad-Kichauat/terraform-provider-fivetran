package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var limit = 1000        // REST API response objects limit per HTTP request
const version = "0.2.0" // Current provider version

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key":    {Type: schema.TypeString, Required: true, DefaultFunc: schema.EnvDefaultFunc("FIVETRAN_APIKEY", nil)},
			"api_secret": {Type: schema.TypeString, Required: true, Sensitive: true, DefaultFunc: schema.EnvDefaultFunc("FIVETRAN_APISECRET", nil)},
		},
		ResourcesMap: map[string]*schema.Resource{
			"fivetran_user":        resourceUser(),
			"fivetran_group":       resourceGroup(),
			"fivetran_destination": resourceDestination(),
			"fivetran_connector":   resourceConnector(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"fivetran_user":                dataSourceUser(),
			"fivetran_users":               dataSourceUsers(),
			"fivetran_group":               dataSourceGroup(),
			"fivetran_groups":              dataSourceGroups(),
			"fivetran_group_connectors":    dataSourceGroupConnectors(),
			"fivetran_group_users":         dataSourceGroupUsers(),
			"fivetran_destination":         dataSourceDestination(),
			"fivetran_connectors_metadata": dataSourceConnectorsMetadata(),
			"fivetran_connector":           dataSourceConnector(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	fivetranClient := fivetran.New(d.Get("api_key").(string), d.Get("api_secret").(string))
	fivetranClient.CustomUserAgent("terraform-provider-fivetran/" + version)
	return fivetranClient, diag.Diagnostics{}
}
