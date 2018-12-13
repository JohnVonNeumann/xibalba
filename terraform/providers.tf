# Don't specify any further information regardiing the aws keys or regions and
# simply allow the environment to handle that
provider "aws" {}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {}
