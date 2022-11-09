package mediaconnect

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	mctypes "github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	resourceHelper "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/intf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	registerFrameworkResourceFactory(newResourceEntitlement)
}

func newResourceEntitlement(_ context.Context) (intf.ResourceWithConfigureAndImportState, error) {
	return &entitlement{}, nil
}

func NewResourceEntitlement(_ context.Context) resource.Resource {
	return &entitlement{}
}

const (
	ResNameEntitlement = "Entitlement"
)

type entitlement struct {
	meta *conns.AWSClient
}

func (m *entitlement) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_mediaconnect_entitlement"
}

func (m *entitlement) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"entitlement_arn": {
				Type:     types.StringType,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					resource.UseStateForUnknown(),
				},
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.RequiresReplace(),
				},
			},
			"flow_arn": {
				Type:     types.StringType,
				Required: true,
			},
			"data_transfur_subscriber_fee_percent": {
				Type:     types.Int64Type,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					int64validator.Between(0, 100),
				},
			},
			"description": {
				Type:     types.StringType,
				Required: true,
			},
			"entitlement_status": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					stringvalidator.OneOf(entitlementStatusToSlice(mctypes.EntitlementStatus("").Values())...),
				},
			},
			"subscribers": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
		},
		Blocks: map[string]tfsdk.Block{

			"encryption": {
				NestingMode: tfsdk.BlockNestingModeList,
				MinItems:    1,
				MaxItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"role_arn": {
						Type:     types.StringType,
						Required: true,
					},
					"algorithm": {
						Type: types.StringType,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf(algorithmsToSlice(mctypes.Algorithm("").Values())...),
						},
					},
					"constant_initialization_vector": {
						Type: types.StringType,
					},
					"device_id": {
						Type: types.StringType,
					},
					"key_type": {
						Type: types.StringType,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf(keyTypeToSlice(mctypes.KeyType("").Values())...),
						},
					},
					"region": {
						Type: types.StringType,
					},
					"resource_id": {
						Type: types.StringType,
					},
					"secret_arn": {
						Type: types.StringType,
					},
					"url": {
						Type: types.StringType,
					},
				},
			},
		},
	}

	return schema, nil
}

func (m *entitlement) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		m.meta = v
	}
}

func (m *entitlement) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := m.meta.MediaConnectConn

	var plan resourceEntitlementData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subscribersAsString := make([]string, len(plan.Subscribers.Elems))
	resp.Diagnostics.Append((plan.Subscribers.ElementsAs(ctx, &subscribersAsString, false))...)
	if resp.Diagnostics.HasError() {
		return
	}
	entitlementEncryptionList := make([]encryption, 1)
	resp.Diagnostics.Append(plan.Encryption.ElementsAs(ctx, &entitlementEncryptionList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	entitlementEncryption, err := expandEntitlementEncryption(ctx, entitlementEncryptionList)
	resp.Diagnostics.Append(err...)
	if resp.Diagnostics.HasError() {
		return
	}

	ger := mctypes.GrantEntitlementRequest{
		Subscribers:                      subscribersAsString,
		DataTransferSubscriberFeePercent: int32(plan.DataTransferSubscriberFeePercent.Value),
		Description:                      aws.String(plan.Description.Value),
		Encryption:                       entitlementEncryption,
		EntitlementStatus:                mctypes.EntitlementStatus(plan.EntitlementStatus.Value),
		Name:                             aws.String(plan.Name.Value),
	}

	in := &mediaconnect.GrantFlowEntitlementsInput{
		FlowArn: aws.String(plan.FlowArn.Value),
		Entitlements: []mctypes.GrantEntitlementRequest{ger},
	}

	out, errCreate := conn.GrantFlowEntitlements(ctx, in)

	if errCreate != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionCreating, ResNameEntitlement, plan.Name.String(), nil),
			errCreate.Error(),
		)
		return
	}

	var result resourceEntitlementData

	result.ID = types.String{Value: fmt.Sprintf("%s/%s", programName, multiplexId)}
	result.ProgramName = types.String{Value: aws.ToString(out.Entitlement.ProgramName)}
	result.MultiplexID = types.String{Value: plan.MultiplexID.Value}
	result.EntitlementSettings = flattenEntitlementSettings(out.Entitlement.EntitlementSettings, isStateMuxSet)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *entitlement) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := m.meta.MediaConnectConn

	var state resourceEntitlementData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseEntitlementID(state.ID.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionReading, ResNameEntitlement, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := FindEntitlementByID(ctx, conn, multiplexId, programName)

	if tfresource.NotFound(err) {
		diag.NewWarningDiagnostic(
			"AWS Resource Not Found During Refresh",
			fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original Error: %s", err.Error()),
		)
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionReading, ResNameEntitlement, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	sm := make([]videoSettings, 1)
	attErr := req.State.GetAttribute(ctx, path.Root("entitlement_settings").
		AtListIndex(0).AtName("video_settings"), &sm)

	resp.Diagnostics.Append(attErr...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateMuxIsNull bool
	if len(sm) > 0 {
		if len(sm[0].StatemuxSettings.Elems) == 0 {
			stateMuxIsNull = true
		}
	}
	state.EntitlementSettings = flattenEntitlementSettings(out.EntitlementSettings, stateMuxIsNull)
	state.ProgramName = types.String{Value: aws.ToString(out.ProgramName)}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (m *entitlement) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := m.meta.MediaConnectConn

	var plan resourceEntitlementData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseEntitlementID(plan.ID.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionReading, ResNameEntitlement, plan.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	mps := make([]entitlementSettings, 1)
	resp.Diagnostics.Append(plan.EntitlementSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpSettings, stateMuxIsNull, errExpand := expandEntitlementSettings(ctx, mps)

	resp.Diagnostics.Append(errExpand...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &mediaconnect.UpdateEntitlementInput{
		MultiplexId:         aws.String(multiplexId),
		ProgramName:         aws.String(programName),
		EntitlementSettings: mpSettings,
	}

	_, errUpdate := conn.UpdateEntitlement(ctx, in)

	if errUpdate != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionUpdating, ResNameEntitlement, plan.ProgramName.String(), nil),
			errUpdate.Error(),
		)
		return
	}

	//Need to find multiplex program because output from update does not provide state data
	out, errUpdate := FindEntitlementByID(ctx, conn, multiplexId, programName)

	if errUpdate != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionUpdating, ResNameEntitlement, plan.ProgramName.String(), nil),
			errUpdate.Error(),
		)
		return
	}

	plan.EntitlementSettings = flattenEntitlementSettings(out.EntitlementSettings, stateMuxIsNull)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (m *entitlement) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := m.meta.MediaConnectConn

	var state resourceEntitlementData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	programName, multiplexId, err := ParseEntitlementID(state.ID.Value)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionDeleting, ResNameEntitlement, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteEntitlement(ctx, &mediaconnect.DeleteEntitlementInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.MediaConnect, create.ErrActionDeleting, ResNameEntitlement, state.ProgramName.String(), nil),
			err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (m *entitlement) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *entitlement) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resourceEntitlementData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	mps := make([]entitlementSettings, 1)
	resp.Diagnostics.Append(data.EntitlementSettings.ElementsAs(ctx, &mps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(mps[0].VideoSettings.Elems) > 0 || !mps[0].VideoSettings.IsNull() {
		vs := make([]videoSettings, 1)
		resp.Diagnostics.Append(mps[0].VideoSettings.ElementsAs(ctx, &vs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		statMuxSet := len(vs[0].StatmuxSettings.Elems) > 0
		stateMuxSet := len(vs[0].StatemuxSettings.Elems) > 0

		if statMuxSet && stateMuxSet {
			resp.Diagnostics.AddAttributeError(
				path.Root("entitlement_settings").AtListIndex(0).AtName("video_settings").AtListIndex(0).AtName("statmux_settings"),
				"Conflicting Attribute Configuration",
				"Attribute statmux_settings cannot be configured with statemux_settings.",
			)
		}
	}
}

func FindEntitlementByID(ctx context.Context, conn *mediaconnect.Client, multiplexId, programName string) (*mediaconnect.DescribeEntitlementOutput, error) {
	in := &mediaconnect.DescribeEntitlementInput{
		MultiplexId: aws.String(multiplexId),
		ProgramName: aws.String(programName),
	}
	out, err := conn.DescribeEntitlement(ctx, in)
	if err != nil {
		var nfe *mltypes.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resourceHelper.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandEntitlementEncryption(ctx context.Context, apiObjectList []encryption) (*mctypes.Encryption, diag.Diagnostics) {
	if len(apiObjectList) == 0 {
		return nil, nil
	}
	apiObject := apiObjectList[0]

	e := &mctypes.Encryption{}
	return e, nil
}

func expandEntitlementSettings(ctx context.Context, mps []entitlementSettings) (*mltypes.EntitlementSettings, bool, diag.Diagnostics) {
	if len(mps) == 0 {
		return nil, false, nil
	}

	var stateMuxIsNull bool
	data := mps[0]

	l := &mltypes.EntitlementSettings{
		ProgramNumber:            int32(data.ProgramNumber.Value),
		PreferredChannelPipeline: mltypes.PreferredChannelPipeline(data.PreferredChannelPipeline.Value),
	}

	if len(data.ServiceDescriptor.Elems) > 0 && !data.ServiceDescriptor.IsNull() {
		sd := make([]serviceDescriptor, 1)
		err := data.ServiceDescriptor.ElementsAs(ctx, &sd, false)
		if err.HasError() {
			return nil, false, err
		}

		l.ServiceDescriptor = &mltypes.EntitlementServiceDescriptor{
			ProviderName: aws.String(sd[0].ProviderName.Value),
			ServiceName:  aws.String(sd[0].ServiceName.Value),
		}
	}

	if len(data.VideoSettings.Elems) > 0 && !data.VideoSettings.IsNull() {
		vs := make([]videoSettings, 1)
		err := data.VideoSettings.ElementsAs(ctx, &vs, false)
		if err.HasError() {
			return nil, false, err
		}

		l.VideoSettings = &mltypes.MultiplexVideoSettings{
			ConstantBitrate: int32(vs[0].ConstantBitrate.Value),
		}

		// Deprecated: will be removed in the next major version
		if len(vs[0].StatemuxSettings.Elems) > 0 && !vs[0].StatemuxSettings.IsNull() {
			sms := make([]statmuxSettings, 1)
			err := vs[0].StatemuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, false, err
			}

			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: int32(sms[0].MinimumBitrate.Value),
				MaximumBitrate: int32(sms[0].MaximumBitrate.Value),
				Priority:       int32(sms[0].Priority.Value),
			}
		}

		if len(vs[0].StatmuxSettings.Elems) > 0 && !vs[0].StatmuxSettings.IsNull() {
			stateMuxIsNull = true
			sms := make([]statmuxSettings, 1)
			err := vs[0].StatmuxSettings.ElementsAs(ctx, &sms, false)
			if err.HasError() {
				return nil, false, err
			}

			l.VideoSettings.StatmuxSettings = &mltypes.MultiplexStatmuxVideoSettings{
				MinimumBitrate: int32(sms[0].MinimumBitrate.Value),
				MaximumBitrate: int32(sms[0].MaximumBitrate.Value),
				Priority:       int32(sms[0].Priority.Value),
			}
		}
	}

	return l, stateMuxIsNull, nil
}

var (
	statmuxAttrs = map[string]attr.Type{
		"minimum_bitrate": types.Int64Type,
		"maximum_bitrate": types.Int64Type,
		"priority":        types.Int64Type,
	}

	videoSettingsAttrs = map[string]attr.Type{
		"constant_bitrate":  types.Int64Type,
		"statemux_settings": types.ListType{ElemType: types.ObjectType{AttrTypes: statmuxAttrs}},
		"statmux_settings":  types.ListType{ElemType: types.ObjectType{AttrTypes: statmuxAttrs}},
	}

	serviceDescriptorAttrs = map[string]attr.Type{
		"provider_name": types.StringType,
		"service_name":  types.StringType,
	}

	entitlementSettingsAttrs = map[string]attr.Type{
		"program_number":             types.Int64Type,
		"preferred_channel_pipeline": types.StringType,
		"service_descriptor":         types.ListType{ElemType: types.ObjectType{AttrTypes: serviceDescriptorAttrs}},
		"video_settings":             types.ListType{ElemType: types.ObjectType{AttrTypes: videoSettingsAttrs}},
	}
)

func flattenEntitlementSettings(mps *mltypes.EntitlementSettings, stateMuxIsNull bool) types.List {
	elemType := types.ObjectType{AttrTypes: entitlementSettingsAttrs}

	vals := types.Object{AttrTypes: entitlementSettingsAttrs}
	attrs := map[string]attr.Value{}

	if mps == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs["program_number"] = types.Int64{Value: int64(mps.ProgramNumber)}
	attrs["preferred_channel_pipeline"] = types.String{Value: string(mps.PreferredChannelPipeline)}
	attrs["service_descriptor"] = flattenServiceDescriptor(mps.ServiceDescriptor)
	attrs["video_settings"] = flattenVideoSettings(mps.VideoSettings, stateMuxIsNull)

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func flattenServiceDescriptor(sd *mltypes.EntitlementServiceDescriptor) types.List {
	elemType := types.ObjectType{AttrTypes: serviceDescriptorAttrs}

	vals := types.Object{AttrTypes: serviceDescriptorAttrs}
	attrs := map[string]attr.Value{}

	if sd == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs["provider_name"] = types.String{Value: aws.ToString(sd.ProviderName)}
	attrs["service_name"] = types.String{Value: aws.ToString(sd.ServiceName)}

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func flattenStatMuxSettings(mps *mltypes.MultiplexStatmuxVideoSettings) types.List {
	elemType := types.ObjectType{AttrTypes: statmuxAttrs}

	vals := types.Object{AttrTypes: statmuxAttrs}

	if mps == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs := map[string]attr.Value{}
	attrs["minimum_bitrate"] = types.Int64{Value: int64(mps.MinimumBitrate)}
	attrs["maximum_bitrate"] = types.Int64{Value: int64(mps.MaximumBitrate)}
	attrs["priority"] = types.Int64{Value: int64(mps.Priority)}

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func flattenVideoSettings(mps *mltypes.MultiplexVideoSettings, stateMuxIsNull bool) types.List {
	elemType := types.ObjectType{AttrTypes: videoSettingsAttrs}

	vals := types.Object{AttrTypes: videoSettingsAttrs}
	attrs := map[string]attr.Value{}

	if mps == nil {
		return types.List{ElemType: elemType, Elems: []attr.Value{}}
	}

	attrs["constant_bitrate"] = types.Int64{Value: int64(mps.ConstantBitrate)}

	if stateMuxIsNull {
		attrs["statmux_settings"] = flattenStatMuxSettings(mps.StatmuxSettings)
		attrs["statemux_settings"] = types.List{
			Elems:    []attr.Value{},
			ElemType: types.ObjectType{AttrTypes: statmuxAttrs},
		}
	} else {
		attrs["statmux_settings"] = types.List{
			Elems:    []attr.Value{},
			ElemType: types.ObjectType{AttrTypes: statmuxAttrs},
		}
		attrs["statemux_settings"] = flattenStatMuxSettings(mps.StatmuxSettings)
	}

	vals.Attrs = attrs

	return types.List{
		Elems:    []attr.Value{vals},
		ElemType: elemType,
	}
}

func ParseEntitlementID(id string) (programName string, multiplexId string, err error) {
	idParts := strings.Split(id, "/")

	if len(idParts) < 2 || (idParts[0] == "" || idParts[1] == "") {
		err = errors.New("invalid id")
		return
	}

	programName = idParts[0]
	multiplexId = idParts[1]

	return
}

type encryption struct {
	RoleArn                      types.String `tfsdk:"role_arn"`
	Algorithm                    types.String `tfsdk:"algorithm"`
	ConstantInitializationVector types.String `tfsdk:"constant_initialization_vector"`
	DeviceId                     types.String `tfsdk:"device_id"`
	KeyType                      types.String `tfsdk:"key_type"`
	Region                       types.String `tfsdk:"region"`
	ResourceId                   types.String `tfsdk:"resource_id"`
	SecretArn                    types.String `tfsdk:"secret_arn"`
	Url                          types.String `tfsdk:"url"`
}

type resourceEntitlementData struct {
	ID                               types.String `tfsdk:"id"`
	DataTransferSubscriberFeePercent types.Int64  `tfsdk:"data_transfer_subscriber_fee_percent"`
	Description                      types.String `tfsdk:"description"`
	Encryption                       types.List   `tfsdk:"encryption"`
	EntitlementArn                   types.String `tfsdk:"entitlement_arn"`
	EntitlementStatus                types.String `tfsdk:"entitlement_status"`
	FlowArn                          types.String `tfsdk:"flow_arn"`
	Name                             types.String `tfsdk:"name"`
	Subscribers                      types.List   `tfsdk:"subscribers"`
}

func algorithmsToSlice(a []mctypes.Algorithm) []string {
	s := make([]string, 0)
	for _, v := range a {
		s = append(s, string(v))
	}
	return s
}

func entitlementStatusToSlice(e []mctypes.EntitlementStatus) []string {
	s := make([]string, 0)
	for _, v := range e {
		s = append(s, string(v))
	}
	return s
}

func keyTypeToSlice(k []mctypes.KeyType) []string {
	s := make([]string, 0)
	for _, v := range k {
		s = append(s, string(v))
	}
	return s
}
