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
	_ datasource.DataSource              = &hostDataSource{}
	_ datasource.DataSourceWithConfigure = &hostDataSource{}
)

func NewHostDataSource() datasource.DataSource {
	return &hostDataSource{}
}

type hostDataSource struct {
	client *client.Client
}

type hostDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	IPInt       types.String `tfsdk:"ip_int"`
	NetworkID   types.String `tfsdk:"network_id"`
	ClientName  types.String `tfsdk:"client_name"`
	IP          types.String `tfsdk:"ip"`
	Hostname    types.String `tfsdk:"hostname"`
	Description types.String `tfsdk:"description"`
	Site        types.String `tfsdk:"site"`
	Category    types.String `tfsdk:"category"`
	Comment     types.String `tfsdk:"comment"`
	IPVersion   types.String `tfsdk:"ip_version"`
}

func (d *hostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *hostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a GestioIP host by client and IP address.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "GestioIP host identifier.",
				Computed:            true,
			},
			"ip_int": schema.StringAttribute{
				MarkdownDescription: "GestioIP internal IP integer identifier used by the frontend delete flow.",
				Computed:            true,
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "GestioIP network identifier that contains the host IP.",
				Computed:            true,
			},
			"client_name": schema.StringAttribute{
				MarkdownDescription: "GestioIP client name. If omitted, the provider-level client_name is used.",
				Optional:            true,
				Computed:            true,
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "Host IP address.",
				Required:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Host name.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Host description.",
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Host site.",
				Computed:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Host category.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Host comment.",
				Computed:            true,
			},
			"ip_version": schema.StringAttribute{
				MarkdownDescription: "IP version reported by GestioIP.",
				Computed:            true,
			},
		},
	}
}

func (d *hostDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		return
	}

	d.client = providerData.client
}

func (d *hostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the host data source.")
		return
	}

	var config hostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := d.resolveClientName(config.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_name"),
			"Missing GestioIP Client Name",
			"The host data source requires client_name either in the data source or in the provider configuration.",
		)
		return
	}

	host, err := d.client.ReadHost(ctx, clientName, config.IP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read GestioIP Host", err.Error())
		return
	}

	state := hostDataSourceModel{
		ID:          types.StringValue(host.ID),
		IPInt:       types.StringValue(host.IPInt),
		NetworkID:   types.StringValue(host.NetworkID),
		ClientName:  types.StringValue(host.ClientName),
		IP:          types.StringValue(host.IP),
		Hostname:    types.StringValue(host.Hostname),
		Description: types.StringValue(host.Description),
		Site:        types.StringValue(host.Site),
		Category:    types.StringValue(host.Category),
		Comment:     types.StringValue(host.Comment),
		IPVersion:   types.StringValue(host.IPVersion),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *hostDataSource) resolveClientName(dataSourceClientName types.String) (string, bool) {
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
