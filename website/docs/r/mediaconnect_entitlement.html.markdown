---
subcategory: "Elemental MediaConnect"
layout: "aws"
page_title: "AWS: aws_mediaconnect_entitlement"
description: |-
  Terraform resource for managing an AWS Elemental MediaConnect Entitlement.
---

# Resource: aws_mediaconnect_entitlement

Terraform resource for managing an AWS Elemental MediaConnect Entitlement.

## Example Usage

### Basic Usage

```terraform
resource "aws_mediaconnect_entitlement" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Entitlement. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Elemental MediaConnect Entitlement can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_mediaconnect_entitlement.example rft-8012925589
```
