package provider

import (
	"context"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &networkDataSource{}
	_ datasource.DataSourceWithConfigure = &networkDataSource{}
)

func NewNetworkDataSource() datasource.DataSource {
	return &networkDataSource{}
}

type networkDataSource struct {
	client *client.Client
}

type networkDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ClientName  types.String `tfsdk:"client_name"`
	IP          types.String `tfsdk:"ip"`
	Bitmask     types.Int64  `tfsdk:"bitmask"`
	Description types.String `tfsdk:"description"`
	Site        types.String `tfsdk:"site"`
	Category    types.String `tfsdk:"category"`
	Comment     types.String `tfsdk:"comment"`
	Sync        types.Bool   `tfsdk:"sync"`
	IPVersion   types.String `tfsdk:"ip_version"`
}

func (d *networkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *networkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a GestioIP network by client, IP address and bitmask.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "GestioIP internal network identifier.",
				Computed:            true,
			},
			"client_name": schema.StringAttribute{
				MarkdownDescription: "GestioIP client name. If omitted, the provider-level client_name is used.",
				Optional:            true,
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "Network IP address.",
				Required:            true,
			},
			"bitmask": schema.Int64Attribute{
				MarkdownDescription: "Network bitmask.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Network description.",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Network site.",
				Computed:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Network category.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Network comment.",
				Computed:            true,
			},
			"sync": schema.BoolAttribute{
				MarkdownDescription: "Whether network synchronization is enabled in GestioIP.",
				Computed:            true,
			},
			"ip_version": schema.StringAttribute{
				MarkdownDescription: "IP version reported by GestioIP.",
				Computed:            true,
			},
		},
	}
}

func (d *networkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		return
	}

	d.client = providerData.client
}

func (d *networkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the network data source.")
		return
	}

	var config networkDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := d.resolveClientName(config.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_name"),
			"Missing GestioIP Client Name",
			"The network data source requires client_name either in the data source or in the provider configuration.",
		)
		return
	}

	networks, err := d.client.ListNetworks(ctx, clientName)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read GestioIP Network", err.Error())
		return
	}

	network, found := findNetwork(networks, config.IP.ValueString(), config.Bitmask.ValueInt64())
	if !found {
		resp.Diagnostics.AddError(
			"GestioIP Network Not Found",
			"The requested network was not returned by GestioIP for the selected client_name, ip and bitmask.",
		)
		return
	}

	state := networkDataSourceModel{
		ID:          types.StringValue(network.ID),
		ClientName:  types.StringValue(network.ClientName),
		IP:          types.StringValue(network.IP),
		Bitmask:     types.Int64Value(network.Bitmask),
		Description: types.StringValue(network.Description),
		Site:        types.StringValue(network.Site),
		Category:    types.StringValue(network.Category),
		Comment:     types.StringValue(network.Comment),
		Sync:        types.BoolValue(network.Sync),
		IPVersion:   types.StringValue(network.IPVersion),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *networkDataSource) resolveClientName(dataSourceClientName types.String) (string, bool) {
	if !dataSourceClientName.IsNull() && !dataSourceClientName.IsUnknown() && dataSourceClientName.ValueString() != "" {
		return dataSourceClientName.ValueString(), true
	}

	if d.client == nil {
		return "", false
	}

	clientName := d.client.ClientName()
	if clientName == "" {
		return "", false
	}

	return clientName, true
}

func findNetwork(networks []client.Network, ip string, bitmask int64) (client.Network, bool) {
	for _, network := range networks {
		if network.IP == ip && network.Bitmask == bitmask {
			return network, true
		}
	}

	return client.Network{}, false
}
