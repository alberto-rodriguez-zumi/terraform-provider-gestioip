package provider

import (
	"context"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &networkResource{}
	_ resource.ResourceWithConfigure   = &networkResource{}
	_ resource.ResourceWithImportState = &networkResource{}
)

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

type networkResource struct {
	client *client.Client
}

type networkResourceModel struct {
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

func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "GestioIP network resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "GestioIP internal network identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional:            true,
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Network site.",
				Optional:            true,
				Computed:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Network category.",
				Optional:            true,
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Network comment.",
				Optional:            true,
				Computed:            true,
			},
			"sync": schema.BoolAttribute{
				MarkdownDescription: "Whether network synchronization is enabled in GestioIP.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"ip_version": schema.StringAttribute{
				MarkdownDescription: "IP version reported by GestioIP.",
				Computed:            true,
			},
		},
	}
}

func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		return
	}

	r.client = providerData.client
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the network resource.")
		return
	}

	var plan networkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := r.resolveClientName(plan.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_name"),
			"Missing GestioIP Client Name",
			"The network resource requires client_name either in the resource or in the provider configuration.",
		)
		return
	}

	network, err := r.client.CreateNetwork(ctx, client.CreateNetworkInput{
		ClientName:  clientName,
		IP:          plan.IP.ValueString(),
		Bitmask:     plan.Bitmask.ValueInt64(),
		Description: plan.Description.ValueString(),
		Site:        plan.Site.ValueString(),
		Category:    plan.Category.ValueString(),
		Comment:     plan.Comment.ValueString(),
		Sync:        plan.Sync.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create GestioIP Network", err.Error())
		return
	}

	state := networkModelFromAPI(*network)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the network resource.")
		return
	}

	var state networkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := r.client.ReadNetwork(ctx, state.ClientName.ValueString(), state.IP.ValueString(), state.Bitmask.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read GestioIP Network", err.Error())
		return
	}

	newState := networkModelFromAPI(*network)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the network resource.")
		return
	}

	var plan networkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := r.resolveClientName(plan.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_name"),
			"Missing GestioIP Client Name",
			"The network resource requires client_name either in the resource or in the provider configuration.",
		)
		return
	}

	network, err := r.client.UpdateNetwork(ctx, client.UpdateNetworkInput{
		ClientName:  clientName,
		IP:          plan.IP.ValueString(),
		Bitmask:     plan.Bitmask.ValueInt64(),
		Description: plan.Description.ValueString(),
		Site:        plan.Site.ValueString(),
		Category:    plan.Category.ValueString(),
		Comment:     plan.Comment.ValueString(),
		Sync:        plan.Sync.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update GestioIP Network", err.Error())
		return
	}

	newState := networkModelFromAPI(*network)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the network resource.")
		return
	}

	var state networkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNetwork(ctx, client.DeleteNetworkInput{
		ClientName: state.ClientName.ValueString(),
		IP:         state.IP.ValueString(),
		Bitmask:    state.Bitmask.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete GestioIP Network", err.Error())
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError(
		"Import Not Yet Implemented",
		"The gestioip_network resource does not support import yet. A stable import identifier format will be added in a later step.",
	)
}

func (r *networkResource) resolveClientName(resourceClientName types.String) (string, bool) {
	if !resourceClientName.IsNull() && !resourceClientName.IsUnknown() && resourceClientName.ValueString() != "" {
		return resourceClientName.ValueString(), true
	}

	if r.client == nil {
		return "", false
	}

	clientName := r.client.ClientName()
	if clientName == "" {
		return "", false
	}

	return clientName, true
}

func networkModelFromAPI(network client.Network) networkResourceModel {
	return networkResourceModel{
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
}
