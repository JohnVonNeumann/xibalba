output "vpc_id" {
  value = "${aws_vpc.honeypot.id}"
}

output "vpc_cidr" {
  value = "${aws_vpc.honeypot.cidr_block}"
}

output "internet_gateway_id" {
  value = "${aws_internet_gateway.honeypot.id}"
}

output "main_route_table_id" {
  value = "${aws_vpc.honeypot.main_route_table_id}"
}

output "route_tables" {
  value = "${data.aws_route_tables.route_tables.ids}"
}
