# Deploy second end-point
KEY_NAME="sec_endpoint"
KEY_PEM="$KEY_NAME.pem"

# echo "create key pair $KEY_PEM to connect to instances and save it locally"
aws ec2 create-key-pair --key-name $KEY_NAME \
    | jq -r ".KeyMaterial" > $KEY_PEM

# secure the key pair
chmod 400 $KEY_PEM

SEC_GRP="my-sec-grp"

# echo "setup firewall $SEC_GRP"
#aws ec2 create-security-group   \
#    --group-name $SEC_GRP       \
#    --description "Access my instances"

# figure out my ip
MY_IP=$(curl ipinfo.io/ip)
# echo "My IP: $MY_IP"


# echo "setup rule allowing SSH access to $MY_IP only"
aws ec2 authorize-security-group-ingress        \
    --group-name $SEC_GRP --port 22 --protocol tcp \
    --cidr $MY_IP/32

# echo "setup rule allowing HTTP (port 8080) access to $MY_IP only"
aws ec2 authorize-security-group-ingress        \
    --group-name $SEC_GRP --port 8080 --protocol tcp \
    --cidr $MY_IP/32

UBUNTU_20_04_AMI="ami-00aa9d3df94c6c354"

# echo "Creating Ubuntu 20.04 instance..."
RUN_INSTANCES=$(aws ec2 run-instances   \
    --image-id $UBUNTU_20_04_AMI        \
    --instance-type t2.micro            \
    --key-name $KEY_NAME                \
    --security-groups $SEC_GRP)

INSTANCE_ID=$(echo $RUN_INSTANCES | jq -r '.Instances[0].InstanceId')

# echo "Waiting for instance creation..."
# echo $INSTANCE_ID

PUBLIC_IP=$(aws ec2 describe-instances  --instance-ids $INSTANCE_ID |
    jq -r '.Reservations[0].Instances[0].PublicIpAddress'
)

echo "$PUBLIC_IP"

# echo "deploying code to production"
scp -i $KEY_PEM -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" server_app.go ubuntu@$PUBLIC_IP:/home/ubuntu/
scp -i $KEY_PEM -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" worker_app.go ubuntu@$PUBLIC_IP:/home/ubuntu/
scp -i $KEY_PEM -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" worker_setup.sh ubuntu@$PUBLIC_IP:/home/ubuntu/
scp -i $KEY_PEM -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" aws_configure_script.sh ubuntu@$PUBLIC_IP:/home/ubuntu/
scp -i $KEY_PEM -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=60" access_keys.json ubuntu@$PUBLIC_IP:/home/ubuntu/

# echo "setup production environment"
ssh -i $KEY_PEM -o "StrictHostKeyChecking=no" -o "ConnectionAttempts=10" ubuntu@$PUBLIC_IP<<EOF
    sudo apt-get -y update
    sudo apt-get -y install golang
    sudo apt-get -y install tmux
    sudo apt-get install jq
    sudo snap install aws-cli --classic
    go mod init server_app.go
    go get github.com/google/uuid
    tmux new-session -d -s my-session 'go run server_app.go'
    tmux detach-client -s my-session
    exit
EOF

sleep 5

echo "$PUBLIC_IP"