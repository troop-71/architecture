# AWS CDK code for [troop-71.com](https://troop-71.com).

[![Deploy](https://github.com/troop-71/architecture/actions/workflows/deploy.yaml/badge.svg?branch=main)](https://github.com/troop-71/architecture/actions/workflows/deploy.yaml)

## Resources contained herein

- One ecs task
  - 0.5 vCpu
  - 1024 MB mem
  - Running [wiki.js](https://js.wiki/)
- One load balnacer
- One public ip
- One vpc
  - Two subnets
- an A record
  - on the troop-71.com hosted zone
- one rds instance
  -  [T4g micro](https://aws.amazon.com/ec2/instance-types/t4/)
  -  20GB storage
-  Some secrets, security groups, route tables etc
