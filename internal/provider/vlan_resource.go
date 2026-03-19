package provider

import (
	"context"

	"github.com/alberto-rodriguez-zumi/terraform-provider-gestioip/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &vlanResource{}
	_ resource.ResourceWithConfigure   = &vlanResource{}
	_ resource.ResourceWithImportState = &vlanResource{}
)

func NewVLANResource() resource.Resource {
	return &vlanResource{}
}

type vlanResource struct {
	client         *client.Client
	allowOverwrite bool
}

type vlanResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ClientName  types.String `tfsdk:"client_name"`
	Number      types.String `tfsdk:"number"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	BGColor     types.String `tfsdk:"bg_color"`
	FontColor   types.String `tfsdk:"font_color"`
}

func (r *vlanResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (r *vlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "GestioIP VLAN resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "GestioIP VLAN identifier.",
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
			"number": schema.StringAttribute{
				MarkdownDescription: "VLAN number.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "VLAN name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "VLAN description.",
				Optional:            true,
				Computed:            true,
			},
			"bg_color": schema.StringAttribute{
				MarkdownDescription: "VLAN background color used by GestioIP.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("blue"),
			},
			"font_color": schema.StringAttribute{
				MarkdownDescription: "VLAN font color used by GestioIP.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("white"),
			},
		},
	}
}

func (r *vlanResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
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

func (r *vlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the VLAN resource.")
		return
	}

	var plan vlanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := r.resolveClientName(plan.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root("client_name"), "Missing GestioIP Client Name", "The VLAN resource requires client_name either in the resource or in the provider configuration.")
		return
	}

	existingVLAN, err := r.client.ReadVLAN(ctx, clientName, plan.Number.ValueString())
	if err == nil {
		if !r.allowOverwrite {
			resp.Diagnostics.AddError(
				"GestioIP VLAN Already Exists",
				"A VLAN with the same number already exists in GestioIP. Import it into Terraform state or set allow_overwrite = true in the provider configuration.",
			)
			return
		}

		updatedVLAN, err := r.client.UpdateVLAN(ctx, client.UpdateVLANInput{
			ID:          existingVLAN.ID,
			ClientName:  clientName,
			Number:      plan.Number.ValueString(),
			Name:        plan.Name.ValueString(),
			Description: plan.Description.ValueString(),
			BGColor:     plan.BGColor.ValueString(),
			FontColor:   plan.FontColor.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Unable to Overwrite GestioIP VLAN", err.Error())
			return
		}

		state := vlanModelFromAPI(*updatedVLAN)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}
	if !client.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Unable to Check GestioIP VLAN Existence", err.Error())
		return
	}

	vlan, err := r.client.CreateVLAN(ctx, client.CreateVLANInput{
		ClientName:  clientName,
		Number:      plan.Number.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		BGColor:     plan.BGColor.ValueString(),
		FontColor:   plan.FontColor.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create GestioIP VLAN", err.Error())
		return
	}

	state := vlanModelFromAPI(*vlan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the VLAN resource.")
		return
	}

	var state vlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlan, err := r.client.ReadVLAN(ctx, state.ClientName.ValueString(), state.Number.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Read GestioIP VLAN", err.Error())
		return
	}

	newState := vlanModelFromAPI(*vlan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *vlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the VLAN resource.")
		return
	}

	var plan vlanResourceModel
	var state vlanResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientName, ok := r.resolveClientName(plan.ClientName)
	if !ok {
		resp.Diagnostics.AddAttributeError(path.Root("client_name"), "Missing GestioIP Client Name", "The VLAN resource requires client_name either in the resource or in the provider configuration.")
		return
	}

	vlan, err := r.client.UpdateVLAN(ctx, client.UpdateVLANInput{
		ID:          state.ID.ValueString(),
		ClientName:  clientName,
		Number:      plan.Number.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		BGColor:     plan.BGColor.ValueString(),
		FontColor:   plan.FontColor.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update GestioIP VLAN", err.Error())
		return
	}

	newState := vlanModelFromAPI(*vlan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *vlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the VLAN resource.")
		return
	}

	var state vlanResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVLAN(ctx, client.DeleteVLANInput{
		ID:          state.ID.ValueString(),
		ClientName:  state.ClientName.ValueString(),
		Number:      state.Number.ValueString(),
		Name:        state.Name.ValueString(),
		Description: state.Description.ValueString(),
		BGColor:     state.BGColor.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Delete GestioIP VLAN", err.Error())
	}
}

func (r *vlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured GestioIP Client", "The provider client was not configured for the VLAN resource.")
		return
	}

	importClientName, number, err := parseVLANImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid GestioIP VLAN Import ID", err.Error())
		return
	}

	clientName := importClientName
	if clientName == "" {
		clientName = r.client.ClientName()
	}
	if clientName == "" {
		resp.Diagnostics.AddError(
			"Missing GestioIP Client Name",
			"The VLAN import requires client_name in the provider configuration or in the import ID using the format [client_name|]<number>.",
		)
		return
	}

	vlan, err := r.client.ReadVLAN(ctx, clientName, number)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Import GestioIP VLAN", err.Error())
		return
	}

	state := vlanModelFromAPI(*vlan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlanResource) resolveClientName(resourceClientName types.String) (string, bool) {
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

func vlanModelFromAPI(vlan client.VLAN) vlanResourceModel {
	return vlanResourceModel{
		ID:          types.StringValue(vlan.ID),
		ClientName:  types.StringValue(vlan.ClientName),
		Number:      types.StringValue(vlan.Number),
		Name:        types.StringValue(vlan.Name),
		Description: types.StringValue(vlan.Description),
		BGColor:     types.StringValue(vlan.BGColor),
		FontColor:   types.StringValue(vlan.FontColor),
	}
}
