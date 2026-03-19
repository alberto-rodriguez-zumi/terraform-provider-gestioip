package provider

import (
	"context"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &hostResource{}
	_ resource.ResourceWithConfigure   = &hostResource{}
	_ resource.ResourceWithImportState = &hostResource{}
)

func NewHostResource() resource.Resource {
	return &hostResource{}
}

type hostResource struct {
	client         *client.Client
	allowOverwrite bool
}

type hostResourceModel struct {
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

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "GestioIP host resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "GestioIP host identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Host name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Host description.",
				Optional:            true,
				Computed:            true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: "Host site.",
				Required:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Host category.",
				Optional:            true,
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Host comment.",
				Optional:            true,
				Computed:            true,
			},
			"ip_version": schema.StringAttribute{
				MarkdownDescription: "IP version reported by GestioIP.",
				Computed:            true,
			},
		},
	}
}

func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*providerData)
	if !ok {
		return
	}

	r.client = providerData.client
	r.allowOverwrite = providerData.allowOverwrite
}

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the host resource.")
		return
	}

	var plan hostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := r.resolveClientName(plan.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_name"),
			"Missing GestioIP Client Name",
			"The host resource requires client_name either in the resource or in the provider configuration.",
		)
		return
	}

	existingHost, err := r.client.ReadHost(ctx, clientName, plan.IP.ValueString())
	if err == nil {
		if !r.allowOverwrite {
			resp.Diagnostics.AddError(
				"GestioIP Host Already Exists",
				"A host with the same ip already exists in GestioIP. Import it into Terraform state or set allow_overwrite = true in the provider configuration.",
			)
			return
		}

		updatedHost, err := r.client.UpdateHost(ctx, client.UpdateHostInput{
			ID:          existingHost.ID,
			IPInt:       existingHost.IPInt,
			NetworkID:   existingHost.NetworkID,
			ClientName:  clientName,
			IP:          plan.IP.ValueString(),
			Hostname:    plan.Hostname.ValueString(),
			Description: plan.Description.ValueString(),
			Site:        plan.Site.ValueString(),
			Category:    plan.Category.ValueString(),
			Comment:     plan.Comment.ValueString(),
			IPVersion:   existingHost.IPVersion,
		})
		if err != nil {
			resp.Diagnostics.AddError("Unable to Overwrite GestioIP Host", err.Error())
			return
		}

		state := hostModelFromAPI(*updatedHost)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}
	if !client.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to Check GestioIP Host Existence", err.Error())
		return
	}

	host, err := r.client.CreateHost(ctx, client.CreateHostInput{
		ClientName:  clientName,
		IP:          plan.IP.ValueString(),
		Hostname:    plan.Hostname.ValueString(),
		Description: plan.Description.ValueString(),
		Site:        plan.Site.ValueString(),
		Category:    plan.Category.ValueString(),
		Comment:     plan.Comment.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create GestioIP Host", err.Error())
		return
	}

	state := hostModelFromAPI(*host)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the host resource.")
		return
	}

	var state hostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host, err := r.client.ReadHost(ctx, state.ClientName.ValueString(), state.IP.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to Read GestioIP Host", err.Error())
		return
	}

	newState := hostModelFromAPI(*host)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the host resource.")
		return
	}

	var plan hostResourceModel
	var state hostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := r.resolveClientName(plan.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_name"),
			"Missing GestioIP Client Name",
			"The host resource requires client_name either in the resource or in the provider configuration.",
		)
		return
	}

	host, err := r.client.UpdateHost(ctx, client.UpdateHostInput{
		ID:          state.ID.ValueString(),
		IPInt:       state.IPInt.ValueString(),
		NetworkID:   state.NetworkID.ValueString(),
		ClientName:  clientName,
		IP:          plan.IP.ValueString(),
		Hostname:    plan.Hostname.ValueString(),
		Description: plan.Description.ValueString(),
		Site:        plan.Site.ValueString(),
		Category:    plan.Category.ValueString(),
		Comment:     plan.Comment.ValueString(),
		IPVersion:   state.IPVersion.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update GestioIP Host", err.Error())
		return
	}

	newState := hostModelFromAPI(*host)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the host resource.")
		return
	}

	var state hostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteHost(ctx, client.DeleteHostInput{
		IPInt:      state.IPInt.ValueString(),
		NetworkID:  state.NetworkID.ValueString(),
		ClientName: state.ClientName.ValueString(),
		IP:         state.IP.ValueString(),
		IPVersion:  state.IPVersion.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete GestioIP Host", err.Error())
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the host resource.")
		return
	}

	importClientName, ip, err := parseHostImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid GestioIP Host Import ID", err.Error())
		return
	}

	clientName := importClientName
	if clientName == "" {
		clientName = r.client.ClientName()
	}
	if clientName == "" {
		resp.Diagnostics.AddError(
			"Missing GestioIP Client Name",
			"The host import requires client_name in the provider configuration or in the import ID using the format [client_name|]<ip>.",
		)
		return
	}

	host, err := r.client.ReadHost(ctx, clientName, ip)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Import GestioIP Host", err.Error())
		return
	}

	state := hostModelFromAPI(*host)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostResource) resolveClientName(resourceClientName types.String) (string, bool) {
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

func hostModelFromAPI(host client.Host) hostResourceModel {
	return hostResourceModel{
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
}
