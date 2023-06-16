#!/bin/bash

# Set the IP address of the EC2 instance to be removed
instance_id="$1"

# Terminate the EC2 instance
aws ec2 terminate-instances --instance-ids "$instance_id"
