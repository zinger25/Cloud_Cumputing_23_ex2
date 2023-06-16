# debug
# set -o xtrace

ip1="$1"
ip2="$2"

scp -i first_endpoint.pem -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" endpoints.json ubuntu@$ip1:/home/ubuntu/
scp -i sec_endpoint.pem -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" endpoints.json ubuntu@$ip2:/home/ubuntu/