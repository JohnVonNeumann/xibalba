output "vpc_cidr" {
  value = "${aws_vpc.honeypot.cidr_block}"
}
