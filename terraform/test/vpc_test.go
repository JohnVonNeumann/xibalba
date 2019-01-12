package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
    // renamed to aws2 to avoid name collision
    aws2 "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/aws/external"
)

func TestTerraformVpcTemplate(t *testing.T) {
	t.Parallel()

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
    // Issue found with this is you can come across dodgy regions without as much support and fine code breaks, like
    // regions not having enough AZ's to support 3 separate subnets in TF code
	awsRegion := aws.GetRandomStableRegion(t, nil, nil)

	terraformOptions := &terraform.Options{
		TerraformDir: "../../terraform",

		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

    vpcCidr := terraform.Output(t, terraformOptions, "vpc_cidr")
    vpcId := terraform.Output(t, terraformOptions, "vpc_id")
    vpcSubnets := aws.GetSubnetsForVpc(t, vpcId, awsRegion)

    var subnetList []string
    for _, subnet := range vpcSubnets {
        subnetList = append(subnetList, subnet.Id)
    }

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

    var subnetCidrList []string
    for _, subnet := range resp.Subnets {
      subnetCidrList = append(subnetCidrList, *subnet.CidrBlock)
    }

    acceptableCidrList := [2]string{"10.0.0.0/28","10.0.1.0/28"}

    assert.ElementsMatch(t, subnetCidrList, acceptableCidrList)
    assert.Equal(t, vpcCidr, "10.0.0.0/16")
    assert.Equal(t, len(vpcSubnets), 2)
}
