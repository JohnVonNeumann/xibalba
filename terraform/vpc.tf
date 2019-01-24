# VPC Resources
# * VPC
# * Subnets
# * Internet Gateways
# * Route Table

resource "aws_vpc" "honeypot" {
  // should probably source this number from a var file
  cidr_block = "10.0.0.0/16"
  enable_dns_support = "true"
  // https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.htm
  // work out whether we should enable dns_hostnames
  // enable_dns_hostnames = "?"

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
  // TODO implement test to ensure that all public subnets are in fact public
  //  map_public_ip_on_launch = "true"

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

data "aws_route_table" "default_vpc_route_table" {
  vpc_id = "${aws_vpc.honeypot.id}"
}

resource "aws_default_route_table" "honeypot" {
  default_route_table_id = "${data.aws_route_table.default_vpc_route_table.id}"
}

data "aws_route_tables" "route_tables" {
  vpc_id = "${aws_vpc.honeypot.id}"
}

resource "aws_route_table_association" "honeypot" {
  count = 2

  subnet_id      = "${aws_subnet.honeypot.*.id[count.index]}"
  route_table_id = "${aws_default_route_table.honeypot.id}"
}

resource "aws_route" "honeypot" {
  route_table_id = "${aws_default_route_table.honeypot.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.honeypot.id}"

  timeouts {
    create = "10s"
    delete = "10s"
  }
}
