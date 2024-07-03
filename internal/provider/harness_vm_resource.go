package provider

import (
	"context"
	"fmt"

	"github.com/chainguard-dev/terraform-provider-imagetest/internal/harnesses/gce"
	"github.com/chainguard-dev/terraform-provider-imagetest/internal/log"
	"github.com/chainguard-dev/terraform-provider-imagetest/internal/util"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &HarnessMachineResource{}
	_ resource.ResourceWithConfigure   = &HarnessMachineResource{}
	_ resource.ResourceWithImportState = &HarnessMachineResource{}
	_ resource.ResourceWithModifyPlan  = &HarnessMachineResource{}
)

func NewHarnessMachineResource() resource.Resource {
	return &HarnessMachineResource{}
}

// HarnessMachineResource defines the resource implementation.
type HarnessMachineResource struct {
	HarnessResource
}

// HarnessMachineResourceModel describes the resource data model.
type HarnessMachineResourceModel struct {
	Id        types.String             `tfsdk:"id"`
	Name      types.String             `tfsdk:"name"`
	Inventory InventoryDataSourceModel `tfsdk:"inventory"`
	Skipped   types.Bool               `tfsdk:"skipped"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type HarnessMachineGCEResourceModel struct{}

func (r *HarnessMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_harness_machine"
}

func (r *HarnessMachineResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	schemaAttributes := util.MergeSchemaMaps(
		addHarnessResourceSchemaAttributes(ctx),
		map[string]schema.Attribute{})

	resp.Schema = schema.Schema{
		MarkdownDescription: `A harness that runs arbitrary terraform on a given path.`,
		Attributes:          schemaAttributes,
	}
}

func (r *HarnessMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HarnessMachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	skipped := r.ShouldSkip(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Skipped = types.BoolValue(skipped)

	if data.Skipped.ValueBool() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	timeout, diags := data.Timeouts.Create(ctx, defaultHarnessCreateTimeout)
	resp.Diagnostics.Append(diags...)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ctx, err := r.store.Logger(ctx, data.Inventory, "harness_id", data.Id.ValueString(), "harness_name", data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to initialize logger(s)", err.Error())
		return
	}

	id := data.Id.ValueString()

	h, err := gce.NewGce()
	if err != nil {
		resp.Diagnostics.AddError("failed to create gce harness", err.Error())
		return
	}
	r.store.harnesses.Set(id, h)

	log.Debug(ctx, fmt.Sprintf("creating gce harness [%s]", id))

	if _, err := h.Setup()(ctx); err != nil {
		resp.Diagnostics.AddError("failed to setup harness", err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HarnessMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HarnessMachineResourceModel
	baseRead(ctx, &data, req, resp)
}

func (r *HarnessMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HarnessMachineResourceModel
	baseUpdate(ctx, &data, req, resp)
}

func (r *HarnessMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HarnessMachineResourceModel
	baseDelete(ctx, &data, req, resp)
}

func (r *HarnessMachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
