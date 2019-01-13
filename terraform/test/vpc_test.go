package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	// renamed to aws2 to avoid name collision
	aws2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

func TestTerraformVpcTemplate(t *testing.T) {
	t.Parallel()

	terraformDir := test_structure.CopyTerraformFolderToTemp(t, "../../", "terraform")

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		terraform.Destroy(t, terraformOptions)
	})

	test.structure.RunTestStage(t, "setup", func() {
		terraformOptions := createTerraformOptions(t, terraformDir)
		test_structure.SaveTerraformOptions(t, terraformDir, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test VPC cidr is acceptable", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testVpcCidrIsCorrect(t, terraformOptions)
	})

	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	igwID := terraform.Output(t, terraformOptions, "internet_gateway_id")
	vpcSubnets := aws.GetSubnetsForVpc(t, vpcID, awsRegion)

	var subnetList []string
	for _, subnet := range vpcSubnets {
		subnetList = append(subnetList, subnet.Id)
	}

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

	subnetParams := &ec2.DescribeSubnetsInput{
		Filters: []ec2.Filter{
			{
				Name:   aws2.String("subnet-id"),
				Values: subnetList,
			},
		},
	}

	subnetReq := client.DescribeSubnetsRequest(subnetParams)

	subnetResp, _ := subnetReq.Send()

	var subnetCidrList []string
	for _, subnet := range subnetResp.Subnets {
		subnetCidrList = append(subnetCidrList, *subnet.CidrBlock)
	}

	igwList := []string{igwID}
	igwParams := &ec2.DescribeInternetGatewaysInput{
		Filters: []ec2.Filter{
			{
				Name:   aws2.String("internet-gateway-id"),
				Values: igwList,
			},
		},
	}

	igwReq := client.DescribeInternetGatewaysRequest(igwParams)
	igwResp, _ := igwReq.Send()

	// test that igw has attachments in available state

	// test that the vpc is the correct vpc

	// test that the vpc has an internet gateway

	fmt.Println(igwResp)

	acceptableCidrList := [2]string{"10.0.0.0/28", "10.0.1.0/28"}

	assert.ElementsMatch(t, subnetCidrList, acceptableCidrList)
	assert.Equal(t, len(vpcSubnets), 2)
}

func createTerraformOptions(t *testing.T, terraformDir string) (*terraform.Options, *aws.Ec2Keypair) {

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

	return terraformOptions

}

// test: testVpcCidrIsCorrect
// assert that the user configured vpcCidr within the terraform template is
// an acceptable value
func testVpcCidrIsCorrect(t *testing.T, terraformOptions *terraformOptions) {
	vpcCidr := terraform.Output(t, terraformOptions, "vpc_cidr")

	assert.Equal(t, vpcCidr, "10.0.0.0/16")
}
