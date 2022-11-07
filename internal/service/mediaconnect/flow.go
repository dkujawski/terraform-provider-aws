package mediaconnect

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceFlow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFlowCreate,
		ReadWithoutTimeout:   resourceFlowRead,
		UpdateWithoutTimeout: resourceFlowUpdate,
		DeleteWithoutTimeout: resourceFlowDelete,

		// TIP: ==== TERRAFORM IMPORTING ====
		// If Read can get all the information it needs from the Identifier
		// (i.e., d.Id()), you can use the Passthrough importer. Otherwise,
		// you'll need a custom import function.
		//
		// See more:
		// https://hashicorp.github.io/terraform-provider-aws/add-import-support/
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#implicit-state-passthrough
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#virtual-attributes
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// TIP: ==== CONFIGURABLE TIMEOUTS ====
		// Users can configure timeout lengths but you need to use the times they
		// provide. Access the timeout they configure (or the defaults) using,
		// e.g., d.Timeout(schema.TimeoutCreate) (see below). The times here are
		// the defaults if they don't configure timeouts.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"maintenance": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maintenance_day": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.MaintenanceDay](),
						},
						"maintenance_start_hour": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"mediastream": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attributes": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fmtp": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"channel_order": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"colorimetry": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.Colorimetry](),
												},
												"exact_framerate": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"par": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"range": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.Range](),
												},
												"scan_mode": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.ScanMode](),
												},
												"tcs": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.Tcs](),
												},
											},
										},
									},
									"lang": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"clock_rate": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"media_stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"media_stream_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"media_stream_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.MediaStreamType](),
						},
						"video_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"decryption": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"algorithm": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.Algorithm](),
									},
									"constant_initialization_vector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"device_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"key_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.KeyType](),
									},
									"region": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"resource_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"role_arn": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
									},
									"secret_arn": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
									},
									"url": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"entitlement_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
						},
						"ingest_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_bitrate": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_latency": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_sync_buffer": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"mediastream_source_config": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encoding_name": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.EncodingName](),
									},
									"stream_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"input_configuration": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"input_port": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"interface": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"min_latency": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"protocol": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.Protocol](),
						},
						"sender_control_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"sender_ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_listener_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_listener_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"stream_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_interface_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"white_list_cidr": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"source_failover_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failover_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.FailoverMode](),
						},
						"recovery_window": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"source_priority": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary_source": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"state": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.State](),
						},
					},
				},
			},
			"vpc_interface": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"network_interface_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.NetworkInterfaceType](),
						},
						"role_arn": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
						},
						"security_group_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},

		//CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameFlow = "Flow"
)

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectConn

	in := &mediaconnect.CreateFlowInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		in.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance"); ok && len(v.([]interface{})) > 0 {
		in.Maintenance = expandMaintenance(v.([]interface{}))
	}

	if v, ok := d.GetOk("mediastream"); ok && len(v.([]interface{})) > 0 {
		in.MediaStreams = expandMediaStreams(v.([]interface{}))
	}

	if v, ok := d.GetOk("source"); ok && len(v.([]interface{})) > 0 {
		in.Sources = expandSources(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_failover_config"); ok && len(v.([]interface{})) > 0 {
		in.SourceFailoverConfig = expandSourceFailoverConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_interface"); ok && len(v.([]interface{})) > 0 {
		in.VpcInterfaces = expandVpcInterfaces(v.([]interface{}))
	}

	out, err := conn.CreateFlow(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameFlow, d.Get("name").(string), err)
	}

	if out == nil || out.Flow == nil {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameFlow, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Flow.FlowArn))

	if _, err := waitFlowCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForCreation, ResNameFlow, d.Id(), err)
	}

	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaConnectConn

	out, err := findFlowByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaConnect Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionReading, ResNameFlow, d.Id(), err)
	}

	d.Set("arn", out.FlowArn)
	d.Set("availability_zone", out.AvailabilityZone)

	d.Set("maintenance", flattenMaintenance(out.Maintenance))
	d.Set("mediastream", flattenMediastream(out.MediaStreams))
	d.Set("source", flattenSource(out.Source))
	d.Set("source_failover_config", flattenSourceFailoverConfig(out.SourceFailoverConfig))
	d.Set("vpc_interface", flattenVpcInterface(out.VpcInterfaces))

	return nil
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	update, err := maybeUpdateFlow(d, meta)
	if err != nil {
		return err
	}
	update, err = maybeUpdateFlowSource(d, meta)
	if err != nil {
		return err
	}
	update, err = maybeUpdateFlowTags(d, meta)
	if err != nil {
		return err
	}
	if !update {
		return nil
	}
	return resourceFlowRead(ctx, d, meta)
}

func maybeUpdateFlow(d *schema.ResourceData, meta interface{}) (bool, diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).MediaConnectConn
	update := false

	in := &mediaconnect.UpdateFlowInput{
		FlowArn: aws.String(d.Id()),
	}

	if d.HasChanges("maintenance") {
		in.Maintenance = expandMaintenance(d.Get("maintenance").(*schema.Set).List())
		update = true
	}

	if d.HasChanges("source_failover_config") {
		in.SourceFailoverConfig = expandSourceFailoverConfig(d.Get("source_failover_config").(*schema.Set).List())
		update = true
	}

	if !update {
		return update, nil
	}

	log.Printf("[DEBUG] Updating MediaConnect Flow (%s): %#v", d.Id(), in)
	out, err := conn.UpdateFlow(ctx, in)
	if err != nil {
		return update, create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
	}

	if _, err := waitFlowUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return update, create.DiagError(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, d.Id(), err)
	}
	return update, nil
}

func maybeUpdateFlowSource(d *schema.ResourceData, meta interface{}) (bool, diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).MediaConnectConn
	update := false

	if d.HasChanges("source") {
		sources := expandSources(d.Get("source").(*schema.Set).Set())
		for source := range(sources) {
			in := &mediaconnect.UpdateFlowSourceInput{
				FlowArn: aws.String(d.Id()),
			}


		}
		update = true
	}

	if !update {
		return update, nil
	}

	log.Printf("[DEBUG] Updating MediaConnect Flow (%s): %#v", d.Id(), in)
	out, err := conn.UpdateFlow(ctx, in)
	if err != nil {
		return update, create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameFlow, d.Id(), err)
	}

	if _, err := waitFlowUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return update, create.DiagError(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameFlow, d.Id(), err)
	}
	return update, nil
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a delete input structure
	// 3. Call the AWS delete function
	// 4. Use a waiter to wait for delete to complete
	// 5. Return nil

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaConnectConn

	// TIP: -- 2. Populate a delete input structure
	log.Printf("[INFO] Deleting MediaConnect Flow %s", d.Id())

	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteFlow(ctx, &mediaconnect.DeleteFlowInput{
		Id: aws.String(d.Id()),
	})

	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaConnect, create.ErrActionDeleting, ResNameFlow, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for delete to complete
	if _, err := waitFlowDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForDeletion, ResNameFlow, d.Id(), err)
	}

	// TIP: -- 5. Return nil
	return nil
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.

func waitFlowCreated(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.Flow, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.Flow); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.

func waitFlowUpdated(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.Flow, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.Flow); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.

func waitFlowDeleted(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.Flow, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusFlow(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.Flow); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.

func statusFlow(ctx context.Context, conn *mediaconnect.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findFlowByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.

func findFlowByID(ctx context.Context, conn *mediaconnect.Client, id string) (*types.Flow, error) {
	in := &mediaconnect.DescribeFlowInput{
		FlowArn: aws.String(id),
	}
	out, err := conn.DescribeFlow(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Flow == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Flow, nil
}

func flattenMaintenance(apiObject *types.Maintenance) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MaintenanceDay; v != "" {
		m["sub_field_one"] = v
	}

	if v := apiObject.MaintenanceDeadline; v != nil {
		m["maintenance_deadline"] = aws.ToString(v)
	}

	if v := apiObject.MaintenanceScheduledDate; v != nil {
		m["maintenance_schedule_date"] = aws.ToString(v)
	}

	if v := apiObject.MaintenanceStartHour; v != nil {
		m["maintenance_start_hour"] = aws.ToString(v)
	}

	return m
}

func flattenFmtp(apiObject *types.Fmtp) []fmtp {
	if apiObject == nil {
		return nil
	}
	f := fmtp{
		ChannelOrder:   tftypes.String{Value: aws.ToString(apiObject.ChannelOrder)},
		Colorimetry:    apiObject.Colorimetry,
		ExactFramerate: tftypes.String{Value: aws.ToString(apiObject.ExactFramerate)},
		Par:            tftypes.String{Value: aws.ToString(apiObject.Par)},
		Range:          apiObject.Range,
		ScanMode:       apiObject.ScanMode,
		Tcs:            apiObject.Tcs,
	}
	return []fmtp{f}
}

func flattenMediaStreamAttributes(apiObject *types.MediaStreamAttributes) []mediaStreamAttributes {
	if apiObject == nil {
		return nil
	}
	attrs := mediaStreamAttributes{
		Fmtp: flattenFmtp(apiObject.Fmtp),
	}
	if v := apiObject.Lang; v != nil {
		attrs.Lang = tftypes.String{Value: aws.ToString(v)}
	}
	return []mediaStreamAttributes{attrs}
}

func flattenMediastream(apiObjects []types.MediaStream) []mediaStream {
	if len(apiObjects) == 0 {
		return nil
	}
	var tfList []mediaStream
	for _, apiObject := range apiObjects {
		if apiObject == (types.MediaStream{}) {
			continue
		}
		ms := mediaStream{
			Fmt:             apiObject.Fmt,
			MediaStreamId:   apiObject.MediaStreamId,
			MediaStreamName: tftypes.String{Value: aws.ToString(apiObject.MediaStreamName)},
			MediaStreamType: apiObject.MediaStreamType,
			Attributes:      flattenMediaStreamAttributes(apiObject.Attributes),
		}
		tfList = append(tfList, ms)
	}
	return tfList
}

func flattenSource(apiObjects []types.Source) []source {

}

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
func flattenComplexArguments(apiObjects []*mediaconnect.ComplexArgument) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}

func expandMaintenance(tfList []interface{}) *types.AddMaintenance {
	if len(tfList) == 0 {
		return nil
	}

	m := &types.AddMaintenance{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["maintenance_day"].(string); ok && v != "" {
		m.MaintenanceDay = types.MaintenanceDay(v)
	}

	if v, ok := tfMap["maintenance_start_hour"].(string); ok && v != "" {
		m.MaintenanceStartHour = aws.String(v)
	}

	return m
}

func expandFmtp(tfMap map[string]interface{}) *types.FmtpRequest {
	f := &types.FmtpRequest{}

	if val, ok := tfMap["channel_order"]; ok {
		f.ChannelOrder = aws.String(val.(string))
	}
	if val, ok := tfMap["colorimetry"]; ok {
		f.Colorimetry = val.(types.Colorimetry)
	}
	if val, ok := tfMap["exact_framerate"]; ok {
		f.ExactFramerate = aws.String(val.(string))
	}
	if val, ok := tfMap["par"]; ok {
		f.Par = aws.String(val.(string))
	}
	if val, ok := tfMap["range"]; ok {
		f.Range = val.(types.Range)
	}
	if val, ok := tfMap["scan_mode"]; ok {
		f.ScanMode = val.(types.ScanMode)
	}
	if val, ok := tfMap["tcs"]; ok {
		f.Tcs = val.(types.Tcs)
	}

	return f
}

func expandMediaStreamAttributes(tfMap map[string]interface{}) *types.MediaStreamAttributesRequest {
	ma := &types.MediaStreamAttributesRequest{}

	if val, ok := tfMap["fmtp"]; ok {
		ma.Fmtp = expandFmtp(val.(map[string]interface{}))
	}

	if val, ok := tfMap["lang"]; ok {
		ma.Lang = aws.String(val.(string))
	}

	return ma
}

func expandMediaStreams(tfList []interface{}) []types.AddMediaStreamRequest {
	if len(tfList) == 0 {
		return nil
	}
	var amsr []types.AddMediaStreamRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		var ms types.AddMediaStreamRequest

		if val, ok := tfMap["attributes"]; ok {
			ms.Attributes = expandMediaStreamAttributes(val.(map[string]interface{}))
		}
		if val, ok := tfMap["clock_rate"]; ok {
			ms.ClockRate = int32(val.(int))
		}
		if val, ok := tfMap["description"]; ok {
			ms.Description = aws.String(val.(string))
		}
		if val, ok := tfMap["media_stream_id"]; ok {
			ms.MediaStreamId = val.(int32)
		}
		if val, ok := tfMap["media_stream_name"]; ok {
			ms.MediaStreamName = aws.String(val.(string))
		}
		if val, ok := tfMap["media_stream_type"]; ok {
			ms.MediaStreamType = val.(types.MediaStreamType)
		}
		if val, ok := tfMap["video_format"]; ok {
			ms.VideoFormat = aws.String(val.(string))
		}

		amsr = append(amsr, ms)

	}

	return amsr
}

func expandInterfaceRequest(tfMap map[string]interface{}) *types.InterfaceRequest {
	return &types.InterfaceRequest{
		Name: aws.String(tfMap["name"].(string)),
	}
}

func expandInputConfigurations(tfList []interface{}) []types.InputConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var icr []types.InputConfigurationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		i := types.InputConfigurationRequest{
			InputPort: int32(tfMap["input_port"].(int)),
			Interface: expandInterfaceRequest(tfMap["interface"].(map[string]interface{})),
		}
		icr = append(icr, i)
	}
	return icr
}
func expandMediaStreamSourceConfigurations(tfList []interface{}) []types.MediaStreamSourceConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var msscr []types.MediaStreamSourceConfigurationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		c := types.MediaStreamSourceConfigurationRequest{
			EncodingName:    tfMap["encoding_name"].(types.EncodingName),
			MediaStreamName: aws.String(tfMap["media_stream_name"].(string)),
		}

		if val, ok := tfMap["input_configurations"]; ok {
			c.InputConfigurations = expandInputConfigurations(val.([]interface{}))
		}

		msscr = append(msscr, c)
	}
	return msscr
}

func expandEncryption(tfMap map[string]interface{}) *types.Encryption {
	e := &types.Encryption{RoleArn: aws.String(tfMap["role_arn"].(string))}

	if val, ok := tfMap["algorithm"]; ok {
		e.Algorithm = val.(types.Algorithm)
	}
	if val, ok := tfMap["constant_initialization_vector"]; ok {
		e.ConstantInitializationVector = aws.String(val.(string))
	}
	if val, ok := tfMap["device_id"]; ok {
		e.DeviceId = aws.String(val.(string))
	}
	if val, ok := tfMap["key_type"]; ok {
		e.KeyType = val.(types.KeyType)
	}
	if val, ok := tfMap["region"]; ok {
		e.Region = aws.String(val.(string))
	}
	if val, ok := tfMap["resource_id"]; ok {
		e.ResourceId = aws.String(val.(string))
	}
	if val, ok := tfMap["secret_arn"]; ok {
		e.SecretArn = aws.String(val.(string))
	}
	if val, ok := tfMap["url"]; ok {
		e.Url = aws.String(val.(string))
	}

	return e
}

func expandSources(tfList []interface{}) []types.SetSourceRequest {
	if len(tfList) == 0 {
		return nil
	}

	var ssr []types.SetSourceRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		var s types.SetSourceRequest
		if val, ok := tfMap["decryption"]; ok {
			s.Decryption = expandEncryption(val.(map[string]interface{}))
		}
		if val, ok := tfMap["description"]; ok {
			s.Description = aws.String(val.(string))
		}
		if val, ok := tfMap["entitlement_arn"]; ok {
			s.EntitlementArn = aws.String(val.(string))
		}
		if val, ok := tfMap["ingest_port"]; ok {
			s.IngestPort = int32(val.(int))
		}
		if val, ok := tfMap["max_bitrate"]; ok {
			s.MaxBitrate = int32(val.(int))
		}
		if val, ok := tfMap["max_latency"]; ok {
			s.MaxLatency = int32(val.(int))
		}
		if val, ok := tfMap["max_sync_buffer"]; ok {
			s.MaxSyncBuffer = int32(val.(int))
		}
		if val, ok := tfMap["media_stream_source_configurations"]; ok {
			s.MediaStreamSourceConfigurations = expandMediaStreamSourceConfigurations(val.([]interface{}))
		}
		if val, ok := tfMap["min_latency"]; ok {
			s.MinLatency = int32(val.(int))
		}
		if val, ok := tfMap["name"]; ok {
			s.Name = aws.String(val.(string))
		}
		if val, ok := tfMap["protocol"]; ok {
			s.Protocol = val.(types.Protocol)
		}
		if val, ok := tfMap["sender_control_port"]; ok {
			s.SenderControlPort = int32(val.(int))
		}
		if val, ok := tfMap["sender_ip_address"]; ok {
			s.SenderIpAddress = aws.String(val.(string))
		}
		if val, ok := tfMap["source_listener_address"]; ok {
			s.SourceListenerAddress = aws.String(val.(string))
		}
		if val, ok := tfMap["source_listener_port"]; ok {
			s.SourceListenerPort = int32(val.(int))
		}
		if val, ok := tfMap["stream_id"]; ok {
			s.StreamId = aws.String(val.(string))
		}
		if val, ok := tfMap["vpc_interface_name"]; ok {
			s.VpcInterfaceName = aws.String(val.(string))
		}
		if val, ok := tfMap["white_list_cidr"]; ok {
			s.WhitelistCidr = aws.String(val.(string))
		}

		ssr = append(ssr, s)

	}
	return ssr
}

func expandSourcePriority(tfMap map[string]interface{}) *types.SourcePriority {
	return &types.SourcePriority{
		PrimarySource: aws.String(tfMap["primary_source"].(string)),
	}
}

func expandSourceFailoverConfig(tfList []interface{}) *types.FailoverConfig {
	if len(tfList) == 0 {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})
	fc := &types.FailoverConfig{}
	if val, ok := tfMap["failover_mode"]; ok {
		fc.FailoverMode = val.(types.FailoverMode)
	}
	if val, ok := tfMap["recovery_window"]; ok {
		fc.RecoveryWindow = int32(val.(int))
	}
	if val, ok := tfMap["source_priority"]; ok {
		fc.SourcePriority = expandSourcePriority(val.(map[string]interface{}))
	}
	if val, ok := tfMap["state"]; ok {
		fc.State = val.(types.State)
	}
	return fc
}

func expandVpcInterfaces(tfList []interface{}) []types.VpcInterfaceRequest {
	if len(tfList) == 0 {
		return nil
	}
	var vir []types.VpcInterfaceRequest
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		v := types.VpcInterfaceRequest{
			Name:             aws.String(tfMap["name"].(string)),
			RoleArn:          aws.String(tfMap["role_arn"].(string)),
			SecurityGroupIds: tfMap["security_group_ids"].([]string),
			SubnetId:         aws.String(tfMap["subnet_id"].(string)),
		}

		if val, ok := tfMap["network_interface_type"]; ok {
			v.NetworkInterfaceType = val.(types.NetworkInterfaceType)
		}
		vir = append(vir, v)
	}
	return vir
}

type mediaStream struct {
	Attributes      []mediaStreamAttributes `tfsdk:"attributes,omitempty"`
	ClockRate       int32                   `tfsdk:"clock_rate,omitempty"`
	Description     tftypes.String          `tfsdk:"description,omitempty"`
	Fmt             int32                   `tfsdk:"fmt,omitempty"`
	MediaStreamId   int32                   `tfsdk:"media_stream_id,omitempty"`
	M:wa
	ediaStreamName tftypes.String          `tfsdk:"media_stream_name,omitempty"`
	MediaStreamType types.MediaStreamType   `tfsdk:"media_stream_type,omitempty"`
	VideoFormat     tftypes.String          `tfsdk:"video_format,omitempty"`
}

type mediaStreamAttributes struct {
	Fmtp []fmtp         `tfsdk:"fmtp,omitempty"`
	Lang tftypes.String `tfsdk:"lang,omitempty"`
}

type fmtp struct {
	ChannelOrder   tftypes.String    `tfsdk:"channel_order,omitempty"`
	Colorimetry    types.Colorimetry `tfsdk:"colorimetry,omitempty"`
	ExactFramerate tftypes.String    `tfsdk:"exact_framerate,omitempty"`
	Par            tftypes.String    `tfsdk:"par,omitempty"`
	Range          types.Range       `tfsdk:"range,omitempty"`
	ScanMode       types.ScanMode    `tfsdk:"scan_mode,omitempty"`
	Tcs            types.Tcs         `tfsdk:"tcs,omitempty"`
}

type source struct {

}
