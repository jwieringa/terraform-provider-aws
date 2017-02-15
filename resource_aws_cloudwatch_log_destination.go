package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsCloudWatchLogDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchLogDestinationPut,
		Update: resourceAwsCloudWatchLogDestinationPut,

		Read:   resourceAwsCloudWatchLogDestinationRead,
		Delete: resourceAwsCloudWatchLogDestinationDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"role_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"target_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudWatchLogDestinationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	name := d.Get("name").(string)
	role_arn := d.Get("role_arn").(string)
	target_arn := d.Get("target_arn").(string)

	params := &cloudwatchlogs.PutDestinationInput{
		DestinationName: aws.String(name),
		RoleArn:         aws.String(role_arn),
		TargetArn:       aws.String(target_arn),
	}

	return resource.Retry(30*time.Second, func() *resource.RetryError {
		resp, err := conn.PutDestination(params)

		if err == nil {
			d.SetId(*resp.Destination.Arn)
			d.Set("arn", *resp.Destination.Arn)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return resource.RetryableError(err)
		}

		if awsErr.Code() == "InvalidParameterException" {
			return resource.NonRetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
}

func resourceAwsCloudWatchLogDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	name := d.Get("name").(string)

	params := &cloudwatchlogs.DescribeDestinationsInput{
		DestinationNamePrefix: aws.String(name),
	}

	resp, err := conn.DescribeDestinations(params)
	if err != nil {
		return fmt.Errorf("Error reading Destinations with name prefix %s: %#v", name, err)
	}

	for _, destination := range resp.Destinations {
		if *destination.DestinationName == name {
			d.SetId(*destination.Arn)
			d.Set("arn", *destination.Arn)
			d.Set("role_arn", *destination.RoleArn)
			d.Set("target_arn", *destination.TargetArn)
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceAwsCloudWatchLogDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	name := d.Get("name").(string)

	params := &cloudwatchlogs.DeleteDestinationInput{
		DestinationName: aws.String(name),
	}
	_, err := conn.DeleteDestination(params)
	if err != nil {
		return fmt.Errorf("Error deleting Destination with name %s", name)
	}
	d.SetId("")
	return nil
}
