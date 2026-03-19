package provider

import (
	"context"
	"fmt"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &gestioIPProvider{}

func New() provider.Provider {
	return &gestioIPProvider{}
}

type gestioIPProvider struct{}

type gestioIPProviderModel struct {
	BaseURL        types.String `tfsdk:"base_url"`
	ClientName     types.String `tfsdk:"client_name"`
	AllowOverwrite types.Bool   `tfsdk:"allow_overwrite"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
}

type providerData struct {
	client         *client.Client
	allowOverwrite bool
}

func allowOverwriteValue(value types.Bool) bool {
	return !value.IsNull() && !value.IsUnknown() && value.ValueBool()
}

func (p *gestioIPProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gestioip"
}

func (p *gestioIPProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for GestioIP.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for the GestioIP API.",
				Required:            true,
			},
			"client_name": schema.StringAttribute{
				MarkdownDescription: "Default GestioIP client name used by resources and data sources that operate within a client context.",
				Optional:            true,
			},
			"allow_overwrite": schema.BoolAttribute{
				MarkdownDescription: "Whether resources should overwrite existing GestioIP objects with the same identity during creation. Defaults to false.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username used to authenticate against GestioIP.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password used to authenticate against GestioIP.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *gestioIPProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data gestioIPProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown GestioIP API Base URL",
			"The provider cannot create the GestioIP client because the base_url value is unknown.",
		)
	}

	if data.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown GestioIP Username",
			"The provider cannot create the GestioIP client because the username value is unknown.",
		)
	}

	if data.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown GestioIP Password",
			"The provider cannot create the GestioIP client because the password value is unknown.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := client.New(client.Config{
		BaseURL:    data.BaseURL.ValueString(),
		ClientName: data.ClientName.ValueString(),
		Username:   data.Username.ValueString(),
		Password:   data.Password.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create GestioIP Client",
			fmt.Sprintf("An unexpected error occurred when creating the GestioIP client: %s", err),
		)
		return
	}

	tflog.Info(ctx, "Configured GestioIP client", map[string]any{
		"base_url":        client.BaseURL(),
		"client_name":     client.ClientName(),
		"allow_overwrite": allowOverwriteValue(data.AllowOverwrite),
	})

	providerData := &providerData{
		client:         client,
		allowOverwrite: allowOverwriteValue(data.AllowOverwrite),
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *gestioIPProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHostResource,
		NewNetworkResource,
		NewVLANResource,
	}
}

func (p *gestioIPProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHostDataSource,
		NewNetworkDataSource,
		NewVLANDataSource,
	}
}
