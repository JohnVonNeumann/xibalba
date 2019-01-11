package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
//	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
    // renamed to aws2 to avoid name collision
    aws2 "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/aws/external"
)

// An example of how to test the Terraform module in examples/terraform-aws-example using Terratest.
func TestTerraformAwsExample(t *testing.T) {
    // no ret val
	t.Parallel()

	// Give this EC2 Instance a unique ID for a name tag so we can distinguish it from any other EC2 Instance running
	// in your AWS account
	// expectedName := fmt.Sprintf("terratest-aws-example-%s", random.UniqueId())

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
    // Issue found with this is you can come across dodgy regions without as much support and fine code breaks, like
    // regions not having enough AZ's to support 3 separate subnets in TF code
	awsRegion := aws.GetRandomStableRegion(t, nil, nil)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../../terraform",

		// Variables to pass to our Terraform code using -var options
		// Vars: map[string]interface{}{
		// 	"instance_name": expectedName,
		// },

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of an output variable
    vpcCidr := terraform.Output(t, terraformOptions, "vpc_cidr")
    vpcId := terraform.Output(t, terraformOptions, "vpc_id")
    vpcSubnets := aws.GetSubnetsForVpc(t, vpcId, awsRegion)

    // subnetList must be created from vpcSubnets as vpcSubnets returns a struct
    // with more information than is required, in particular, it included the
    // availability zones of the subnets.
    var subnetList []string
    for _, subnet := range vpcSubnets {
        subnetList = append(subnetList, subnet.Id)
    }

    fmt.Printf("listing subnets belonging to %v : %v\n", vpcId, subnetList)

    cfg, _ := external.LoadDefaultAWSConfig()
    cfg.Region = awsRegion
    client := ec2.New(cfg)

    params := &ec2.DescribeSubnetsInput{
            Filters: []ec2.Filter{
                    {
                            Name: aws2.String("subnet-id"),
                            Values: subnetList,
                    },
            },
    }

    req := client.DescribeSubnetsRequest(params)

    resp, _ := req.Send()

    fmt.Printf("%+v\n", *resp)

    // traverse the json resp, finding only the subnet cidr block and
    // append it to an array/slice much like we did with the subnetList
    // var subnetCidrList []string
//    for index, subnet := range resp {
//      //  subnetCidrList = append(subnetCidrList, subnet.Subnets)
//      fmt.Printf(resp[index].Subnets)
//    }

    fmt.Printf("%+v\n", resp.Subnets)

    // output the array and test against expected

    assert.Equal(t, vpcCidr, "10.0.0.0/16")
    assert.Equal(t, len(vpcSubnets), 2)
    // look for the ip's we expect in the vpcSubnets that we iterate through
    // assert.Contains(t, ["10.0.1.0", "10.0.2.0"]
}
