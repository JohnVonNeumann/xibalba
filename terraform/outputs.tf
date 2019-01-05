output "vpc_id" {
  value = "${aws_vpc.honeypot.id}"
}

output "vpc_cidr" {
  value = "${aws_vpc.honeypot.cidr_block}"
}
