package mediaconnect

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect/types"
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

func ResourceOutput() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOutputCreate,
		ReadWithoutTimeout:   resourceOutputRead,
		UpdateWithoutTimeout: resourceOutputUpdate,
		DeleteWithoutTimeout: resourceOutputDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"cidr_allow_list": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encryption": {
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
			"flow_arn": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
			},
			"max_latency": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"media_stream_output_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_configuration": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_ip": {
										Type:     schema.TypeString,
										Required: true,
									},
									"destination_port": {
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
						"encoding_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.EncodingName](),
						},
						"encoding_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compression_factor": {
										Type:     schema.TypeFloat,
										Required: true,
									},
									"encoder_profile": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.EncoderProfile](),
									},
								},
							},
						},
						"media_stream_name": {
							Type:     schema.TypeString,
							Required: true,
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
			"port": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
			},
			"remote_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sender_control_port": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"smoothing_latency": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"stream_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpc_interface_attachment": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_interface_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},

		//CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameOutput = "Output"
)

func resourceOutputCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	aor := types.AddOutputRequest{Protocol: d.Get("protocol").(types.Protocol)}

	if v, ok := d.GetOk("description"); ok {
		aor.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("destination"); ok {
		aor.Destination = aws.String(v.(string))
	}
	if v, ok := d.GetOk("name"); ok {
		aor.Name = aws.String(v.(string))
	}
	if v, ok := d.GetOk("remote_id"); ok {
		aor.RemoteId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("stream_id"); ok {
		aor.StreamId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maxLatency"); ok {
		aor.MaxLatency = v.(int32)
	}
	if v, ok := d.GetOk("minLatency"); ok {
		aor.MinLatency = v.(int32)
	}
	if v, ok := d.GetOk("port"); ok {
		aor.Port = v.(int32)
	}
	if v, ok := d.GetOk("senderControlPort"); ok {
		aor.SenderControlPort = v.(int32)
	}
	if v, ok := d.GetOk("smoothingLatency"); ok {
		aor.SmoothingLatency = v.(int32)
	}

	if v, ok := d.GetOk("cidr_allow_list"); ok {
		aor.CidrAllowList = v.([]string)
	}

	if v, ok := d.GetOk("encryption"); ok {
		aor.Encryption = expandEncryption(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("media_stream_output_configuration"); ok {
		aor.MediaStreamOutputConfigurations = expandMediaStreamOutputConfigurations(v.([]interface{}))
	}
	if v, ok := d.GetOk("vpc_interface_attachment"); ok {
		aor.VpcInterfaceAttachment = &types.VpcInterfaceAttachment{
			VpcInterfaceName: aws.String(v.(map[string]interface{})["vpc_interface_name"].(string)),
		}
	}

	in := &mediaconnect.AddFlowOutputsInput{
		FlowArn: aws.String(d.Get("flow_arn").(string)),
		Outputs: []types.AddOutputRequest{aor},
	}

	conn := meta.(*conns.AWSClient).MediaConnectConn
	out, err := conn.AddFlowOutputs(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameOutput, d.Get("name").(string), err)
	}

	if out == nil || len(out.Outputs) == 0 {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameOutput, d.Get("name").(string), errors.New("empty output"))
	}

	for _, output := range out.Outputs {
		if aor.Name == output.Name {
			d.SetId(aws.ToString(output.OutputArn))
			break
		}
	}
	if d.Id() == "" {
		return create.DiagError(names.MediaConnect, create.ErrActionCreating, ResNameOutput, d.Get("name").(string), errors.New("empty output"))
	}

	if _, err := waitOutputCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForCreation, ResNameOutput, d.Id(), err)
	}

	return resourceOutputRead(ctx, d, meta)
}

func resourceOutputRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get the resource from AWS
	// 3. Set ID to empty where resource is not new and not found
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return nil

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaConnectConn

	// TIP: -- 2. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findOutputByID(ctx, conn, d.Id())

	// TIP: -- 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaConnect Output (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionReading, ResNameOutput, d.Id(), err)
	}

	// TIP: -- 4. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.TypeString, schema.TypeBool,
	// schema.TypeInt, and schema.TypeFloat), a simple Set call (e.g.,
	// d.Set("arn", out.Arn) is sufficient. No error or nil checking is
	// necessary.
	//
	// However, there are some situations where more handling is needed.
	// a. Complex data types (e.g., schema.TypeList, schema.TypeSet)
	// b. Where errorneous diffs occur. For example, a schema.TypeString may be
	//    a JSON. AWS may return the JSON in a slightly different order but it
	//    is equivalent to what is already set. In that case, you may check if
	//    it is equivalent before setting the different JSON.
	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#data-handling-and-conversion
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#flatten-functions-for-blocks
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#root-typeset-of-resource-and-aws-list-of-structure
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameOutput, d.Id(), err)
	}

	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameOutput, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionSetting, ResNameOutput, d.Id(), err)
	}

	d.Set("policy", p)

	// TIP: -- 6. Return nil
	return nil
}

func resourceOutputUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have ForceNew: true, set
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the main resource function will not have a
	// UpdateWithoutTimeout defined. In the case of c., Update and Create are
	// the same.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a modify input structure and check for changes
	// 3. Call the AWS modify/update function
	// 4. Use a waiter to wait for update to complete
	// 5. Call the Read function in the Update return

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaConnectConn

	// TIP: -- 2. Populate a modify input structure and check for changes
	//
	// When creating the input structure, only include mandatory fields. Other
	// fields are set as needed. You can use a flag, such as update below, to
	// determine if a certain portion of arguments have been changed and
	// whether to call the AWS update function.
	update := false

	in := &mediaconnect.UpdateOutputInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return nil. Otherwise, return a read call, as below.
		return nil
	}

	// TIP: -- 3. Call the AWS modify/update function
	log.Printf("[DEBUG] Updating MediaConnect Output (%s): %#v", d.Id(), in)
	out, err := conn.UpdateOutput(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionUpdating, ResNameOutput, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for update to complete
	if _, err := waitOutputUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForUpdate, ResNameOutput, d.Id(), err)
	}

	// TIP: -- 5. Call the Read function in the Update return
	return resourceOutputRead(ctx, d, meta)
}

func resourceOutputDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	log.Printf("[INFO] Deleting MediaConnect Output %s", d.Id())

	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteOutput(ctx, &mediaconnect.DeleteOutputInput{
		Id: aws.String(d.Id()),
	})

	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaConnect, create.ErrActionDeleting, ResNameOutput, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for delete to complete
	if _, err := waitOutputDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaConnect, create.ErrActionWaitingForDeletion, ResNameOutput, d.Id(), err)
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

func waitOutputCreated(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.Output, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusOutput(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.Output); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.

func waitOutputUpdated(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.Output, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusOutput(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.Output); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.

func waitOutputDeleted(ctx context.Context, conn *mediaconnect.Client, id string, timeout time.Duration) (*mediaconnect.Output, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusOutput(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediaconnect.Output); ok {
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

func statusOutput(ctx context.Context, conn *mediaconnect.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findOutputByID(ctx, conn, id)
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

func findOutputByID(ctx context.Context, conn *mediaconnect.Client, id string) (*mediaconnect.Output, error) {
	in := &mediaconnect.GetOutputInput{
		Id: aws.String(id),
	}
	out, err := conn.GetOutput(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Output == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Output, nil
}

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return something you can use in
// a d.Set() call. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func flattenComplexArgument(apiObject *mediaconnect.ComplexArgument) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
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

func expandDestinationConfigurations(tfList []interface{}) []types.DestinationConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}
	var dcr []types.DestinationConfigurationRequest
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		r := types.DestinationConfigurationRequest{
			DestinationIp:   aws.String(tfMap["destination_ip"].(string)),
			DestinationPort: tfMap["destination_port"].(int32),
			Interface: &types.InterfaceRequest{
				Name: aws.String(tfMap["interface"].(map[string]interface{})["name"].(string)),
			},
		}
		dcr = append(dcr, r)
	}
	return dcr
}
func expandMediaStreamOutputConfigurations(tfList []interface{}) []types.MediaStreamOutputConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}
	var msocr []types.MediaStreamOutputConfigurationRequest
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		cr := types.MediaStreamOutputConfigurationRequest{
			EncodingName:    tfMap["encoding_name"].(types.EncodingName),
			MediaStreamName: aws.String(tfMap["media_stream_name"].(string)),
		}
		if v, ok := tfMap["destination_configurations"]; ok {
			cr.DestinationConfigurations = expandDestinationConfigurations(v.([]interface{}))
		}
		if v, ok := tfMap["encoding_parameters"]; ok {
			rawParams := v.(map[string]interface{})
			cr.EncodingParameters = &types.EncodingParametersRequest{
				CompressionFactor: rawParams["compression_factor"].(float64),
				EncoderProfile:    rawParams["encoder_profile"].(types.EncoderProfile),
			}
		}
		msocr = append(msocr, cr)
	}
	return msocr
}
