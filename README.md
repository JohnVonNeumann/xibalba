# Xibalba

> Xibalba (Mayan pronunciation: [ʃiɓalˈɓa]), roughly translated as "place of fear", is the name of the underworld in K'iche' Maya mythology, ruled by the Maya death gods and their helpers.

Xibalba is a simple project for hosting an SSH honeypot, the hosts will be open to the world on `0.0.0.0:22` and all failed attempts to access the host will be recorded in an attempt to create a usable dataset. The project will be supported by a solid DevOps process, where nightly rebuilds of the host will be the norm, in order to ensure that the box stays secure and well supported. This is an attempt to not only learn more about DevOps, but also DevSecOps, an area I have a lot of interest in.

I am streaming the development of this repository over at Twitch, to see times and other information, visit:
https://www.twitch.tv/janosvonneumann

## To do
* Implement CI tooling around the repository to automate test runs, linting, blah
* Git hooks for better local environment level testing/linting
* Set up host level security
* Set up bake and provision pipeline
* Set up monitoring/logging
* Set up automated pentesting/recon and attacking via separate boxes
* Set up vulnerability scanners
* Set up centralised patch management and audit results
* Set up prowler in build pipeline
* Find an implement an appropriate level of compliance, probably CIS
        - https://d0.awsstatic.com/whitepapers/compliance/AWS_CIS_Foundations_Benchmark.pdf
* Work out and address issues at each level of the Cyber Kill Chain
        - https://nigesecurityguy.wordpress.com/tag/cyber-kill-chain/ 

## Technologies currently/planned to be used in this project

### Currently used:
* Terraform
* Go
* Terratest
* AWS

### Planned for usage:
* Ansible
* Ansible molecule
* Automated Source Code Analysis/Linting
* Continuous Integration
* Inspec
* OSSec
* Security Onion

## Todo
* Work out which SSH honeypot service we wish to use.

## Usage
Make sure the environment is populated with the correct envvars to run the TF.

# Populating the environment with the correct env vars
I would recommend using [awskeyring](https://github.com/vibrato/awskeyring) by
my fantastic team over at [Vibrato](https://github.com/vibrato). RTFM for good
instructions on how to use the tool.
