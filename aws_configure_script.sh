#!/bin/bash

# Path to the JSON file containing the access key parameters
json_file="access_keys.json"

# Read the access key parameters from the JSON file
access_key_id=$(jq -r '.accessKeyId' "$json_file")
secret_access_key=$(jq -r '.secretAccessKey' "$json_file")
region=$(jq -r '.region' "$json_file")

# Configure the AWS CLI with the access key parameters
aws configure set aws_access_key_id "$access_key_id"
aws configure set aws_secret_access_key "$secret_access_key"
aws configure set region "$region"

echo "AWS CLI configured successfully!"