package test

import (
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

	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions := createTerraformOptions(t, terraformDir)
		test_structure.SaveTerraformOptions(t, terraformDir, terraformOptions)

		terraform.InitAndApply(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test VPC cidr is acceptable", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testVpcCidrIsCorrect(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test VPC subnet count", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testVpcSubnetCount(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test subnet cidrs", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testSubnetCidrs(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test igw attachments exist", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testIgwAttachmentsAreAvailable(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test igw attachments vpcId are correct", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testIgwIsAttachingToCorrectVpc(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test route table count for vpc", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testRouteTableCountForVpc(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test rt attachment count is correct", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testVpcRouteTableAssociationCount(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test rt route count is correct", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testRouteTableRouteCount(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test rt is being inherited to stack correctly", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testMainRouteTableId(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test rt contains route to public internet", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testInternetFacingRouteInRouteTable(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test rt contains route to public internet", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testInternetFacingRouteCount(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "test subnets are public and serving ipv4s", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, terraformDir)
		testSubnetsArePublic(t, terraformOptions)
	})

}

func createTerraformOptions(t *testing.T, terraformDir string) *terraform.Options {

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
	// Issue found with this is you can come across dodgy regions without as much support and fine code breaks, like
	// regions not having enough AZ's to support 3 separate subnets in TF code
	awsRegion := aws.GetRandomStableRegion(t, nil, nil)

	terraformOptions := &terraform.Options{
		TerraformDir: "../../terraform",

		Vars: map[string]interface{}{
			"aws_region": awsRegion,
		},

		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": awsRegion,
		},
	}

	return terraformOptions

}

// test: testVpcCidrIsCorrect
// assert that the user configured vpcCidr within the terraform template is
// an acceptable value
func testVpcCidrIsCorrect(t *testing.T, terraformOptions *terraform.Options) { // {{{
	vpcCidr := terraform.Output(t, terraformOptions, "vpc_cidr")

	assert.Equal(t, vpcCidr, "10.0.0.0/16")
} // }}}

// test: testVpcSubnetCount
// assert that the VPC we have created is associated with the correct amount of
// subnets
func testVpcSubnetCount(t *testing.T, terraformOptions *terraform.Options) { // {{{
	awsRegion := terraformOptions.Vars["aws_region"].(string)
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
	vpcSubnets := aws.GetSubnetsForVpc(t, vpcID, awsRegion)

	assert.Equal(t, len(vpcSubnets), 2)
} // }}}

// test: testSubnetCidrs
// assert that the subnet cidrs are within the correct range and that the
// subnet masks are also correct
func testSubnetCidrs(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
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

	acceptableCidrList := [2]string{"10.0.0.0/28", "10.0.1.0/28"}

	assert.ElementsMatch(t, subnetCidrList, acceptableCidrList)
}

// test that igw has attachments in available state
func testIgwAttachmentsAreAvailable(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	igwID := terraform.Output(t, terraformOptions, "internet_gateway_id")

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

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

	var igwAttachmentList []string
	for _, igw := range igwResp.InternetGateways {
		for _, attachment := range igw.Attachments {
			igwAttachmentList = append(igwAttachmentList, string(attachment.State))
		}
	}

	for _, igwAttachment := range igwAttachmentList {
		assert.Equal(t, igwAttachment, "available")
	}

}

// test that the vpc is the correct vpc
func testIgwIsAttachingToCorrectVpc(t *testing.T, terraformOptions *terraform.Options) {
	// {{{
	awsRegion := terraformOptions.Vars["aws_region"].(string)
	igwID := terraform.Output(t, terraformOptions, "internet_gateway_id")
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

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

	var igwAttachmentList []string
	for _, igw := range igwResp.InternetGateways {
		for _, attachment := range igw.Attachments {
			igwAttachmentList = append(igwAttachmentList, *attachment.VpcId)
		}
	}

	for _, igwAttachment := range igwAttachmentList {
		assert.Equal(t, igwAttachment, vpcID)
	}
}

// }}}

// test route table count is 1 via the aws_route_tables data source
// https://www.terraform.io/docs/providers/aws/d/route_tables.html
func testRouteTableCountForVpc(t *testing.T, terraformOptions *terraform.Options) {

	routeTables := terraform.OutputList(t, terraformOptions, "route_tables")

	assert.Len(t, routeTables, 1)
}

// test that there are three route_table_associations for the vpc
// as we will only have two subnets, we should only have three assocs
// with an addtional for the default/main assoc
func testVpcRouteTableAssociationCount(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	routeTableID := terraform.Output(t, terraformOptions, "main_route_table_id")
	// using the attributes, find the number of elements for rt_assocs

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

	rtList := []string{routeTableID}
	rtParams := &ec2.DescribeRouteTablesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws2.String("route-table-id"),
				Values: rtList,
			},
		},
	}

	rtReq := client.DescribeRouteTablesRequest(rtParams)
	rtResp, _ := rtReq.Send()

	var rtAssociationCount int
	for _, rt := range rtResp.RouteTables {
		rtAssociationCount = len(rt.Associations)
	}

	//	// assert that this number is within our acceptance bounds
	assert.Equal(t, 3, rtAssociationCount)

}

// test that the route table has only the amount of routes that we are
// after, im thinking that number is 3, one for the IGW, and two for
// each subnet, however i could be wrong entirely, as i believe that
// any traffic to any ip within the 10.0.0.0/24 range will be routed
// internally as a result of the `local` target dest declaration
// UPDATE: turns out this was incorrect we will only need the single route
// and that will be for the 0.0.0.0 route
// route 1 - dest 0.0.0.0 - target igw
// i think what is happening here, is that the data source/outputs.tf is
// lagging behind the aws api, because i dont seem to get updated info
// out of my outputs in the same way i do from direct aws-sdk calls
// i should probably just bloody learn how the outputs are generated
func testRouteTableRouteCount(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	routeTableID := terraform.Output(t, terraformOptions, "main_route_table_id")

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

	rtList := []string{routeTableID}
	rtParams := &ec2.DescribeRouteTablesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws2.String("route-table-id"),
				Values: rtList,
			},
		},
	}

	rtReq := client.DescribeRouteTablesRequest(rtParams)
	rtResp, _ := rtReq.Send()

	var rtRouteCount int
	for _, rt := range rtResp.RouteTables {
		rtRouteCount = len(rt.Routes)
	}

	assert.Equal(t, 2, rtRouteCount)
}

// test that the route table that we create is the main_route_table_id
// this needs to change in terms of scope, it needs to test that the originally
// created route table is the default and under our control
func testMainRouteTableId(t *testing.T, terraformOptions *terraform.Options) {

	automaticRouteTableId := terraform.Output(t, terraformOptions, "main_route_table_id")
	inheritedRouteTableId := terraform.Output(t, terraformOptions, "route_table_id")

	assert.Equal(t, automaticRouteTableId, inheritedRouteTableId)

}

// test that we have a internet facing route (0.0.0.0/0) in our route table
func testInternetFacingRouteInRouteTable(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	routeTableID := terraform.Output(t, terraformOptions, "main_route_table_id")

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

	rtList := []string{routeTableID}
	rtParams := &ec2.DescribeRouteTablesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws2.String("route-table-id"),
				Values: rtList,
			},
		},
	}

	rtReq := client.DescribeRouteTablesRequest(rtParams)
	rtResp, _ := rtReq.Send()

	var rtRouteDestinations []string
	for _, rt := range rtResp.RouteTables {
		for _, route := range rt.Routes {
			rtRouteDestinations = append(rtRouteDestinations, *route.DestinationCidrBlock)
		}
	}

	assert.Contains(t, rtRouteDestinations, "0.0.0.0/0")

}

// test that there is only one internet facing route
func testInternetFacingRouteCount(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	routeTableID := terraform.Output(t, terraformOptions, "main_route_table_id")

	cfg, _ := external.LoadDefaultAWSConfig()
	cfg.Region = awsRegion
	client := ec2.New(cfg)

	rtList := []string{routeTableID}
	rtParams := &ec2.DescribeRouteTablesInput{
		Filters: []ec2.Filter{
			{
				Name:   aws2.String("route-table-id"),
				Values: rtList,
			},
		},
	}

	rtReq := client.DescribeRouteTablesRequest(rtParams)
	rtResp, _ := rtReq.Send()

	var rtRouteDestinations []string
	for _, rt := range rtResp.RouteTables {
		for _, route := range rt.Routes {

			rtRouteDestinations = append(rtRouteDestinations, *route.DestinationCidrBlock)
		}
	}

	var rtInternetFacingRouteCount int
	for _, destination := range rtRouteDestinations {
		if destination == "0.0.0.0/0" {
			rtInternetFacingRouteCount += 1
		}
	}

	assert.Equal(t, rtInternetFacingRouteCount, 1)
}

// test that the public subnets have the `map_public_ip_on_launch` flag enabled
// this ensures that our instances will be given ipv4 addresses
func testSubnetsArePublic(t *testing.T, terraformOptions *terraform.Options) {

	awsRegion := terraformOptions.Vars["aws_region"].(string)
	vpcID := terraform.Output(t, terraformOptions, "vpc_id")
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

	for _, subnet := range subnetResp.Subnets {
		assert.True(t, *subnet.MapPublicIpOnLaunch)
	}

}
