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
	_ datasource.DataSource              = &vlanDataSource{}
	_ datasource.DataSourceWithConfigure = &vlanDataSource{}
)

func NewVLANDataSource() datasource.DataSource {
	return &vlanDataSource{}
}

type vlanDataSource struct {
	client *client.Client
}

type vlanDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	ClientName  types.String `tfsdk:"client_name"`
	Number      types.String `tfsdk:"number"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	BGColor     types.String `tfsdk:"bg_color"`
	FontColor   types.String `tfsdk:"font_color"`
}

func (d *vlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (d *vlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a GestioIP VLAN by client and number.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"client_name": schema.StringAttribute{
				MarkdownDescription: "GestioIP client name. If omitted, the provider-level client_name is used.",
				Optional:            true,
				Computed:            true,
			},
			"number": schema.StringAttribute{
				MarkdownDescription: "VLAN number.",
				Required:            true,
			},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"bg_color":    schema.StringAttribute{Computed: true},
			"font_color":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *vlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		return
	}
	d.client = providerData.client
}

func (d *vlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the VLAN data source.")
		return
	}

	var config vlanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := d.resolveClientName(config.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root("client_name"), "Missing GestioIP Client Name", "The VLAN data source requires client_name either in the data source or in the provider configuration.")
		return
	}

	vlan, err := d.client.ReadVLAN(ctx, clientName, config.Number.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read GestioIP VLAN", err.Error())
		return
	}

	state := vlanDataSourceModel{
		ID:          types.StringValue(vlan.ID),
		ClientName:  types.StringValue(vlan.ClientName),
		Number:      types.StringValue(vlan.Number),
		Name:        types.StringValue(vlan.Name),
		Description: types.StringValue(vlan.Description),
		BGColor:     types.StringValue(vlan.BGColor),
		FontColor:   types.StringValue(vlan.FontColor),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *vlanDataSource) resolveClientName(dataSourceClientName types.String) (string, bool) {
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
