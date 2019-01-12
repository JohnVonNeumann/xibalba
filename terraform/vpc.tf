# VPC Resources
# * VPC
# * Subnets
# * Internet Gateways
# * Route Table

resource "aws_vpc" "honeypot" {
  // should probably source this number from a var file
  cidr_block = "10.0.0.0/16"

  tags {
    app_id   = "xibalba"
    app_role = "networking"
  }
}

resource "aws_subnet" "honeypot" {
  // originally (and normally) I'd run three subnets because it's typically
  // a better way of handling HA, but the way we are testing this code with
  // terratest spins up in a random region, and some regions only have 2 AZ's
  // so I've reduced down to 2. My main reason for the reduction is that this
  // isn't a critical system, so reduced availability isn't a massive issue.
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/28"
  vpc_id            = "${aws_vpc.honeypot.id}"

  tags {
    app_id   = "xibalba"
    app_role = "networking"
  }
}

resource "aws_internet_gateway" "honeypot" {
  vpc_id = "${aws_vpc.honeypot.id}"

  tags {
    app_id   = "xibalba"
    app_role = "networking"
  }
}

// resource "aws_route_table" "honeypot" {
//   vpc_id = "${aws_vpc.honeypot.id}"
// 
//   route {
//     cidr_block = "0.0.0.0/0"
//     gateway_id = "${aws_internet_gateway.honeypot.id}"
//   }
// 
//   tags {
//     app_id   = "xibalba"
//     app_role = "networking"
//   }
// }
// 
// resource "aws_route_table_association" "honeypot" {
//   count = 3
// 
//   subnet_id      = "${aws_subnet.honeypot.*.id[count.index]}"
//   route_table_id = "${aws_route_table.honeypot.id}"
// }
