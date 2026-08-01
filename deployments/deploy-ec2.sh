#!/bin/bash
# Cloud Guard - one-shot EC2 deploy script
# Run this in AWS CloudShell (console.aws.amazon.com -> CloudShell icon, top nav)
# once your account's "new account verification" clears (usually within ~2 days).
# Region: us-east-1. Requires no extra setup - CloudShell already has AWS CLI + your console credentials.

set -euo pipefail

REGION="us-east-1"
SG_NAME="cloud-guard-sg"
KEY_NAME="cloud-guard-key"
INSTANCE_NAME="cloud-guard-prod"
REPO_URL="https://github.com/singh-anurag-7991/cloud-guard.git"

echo "== Getting default VPC =="
VPC_ID=$(aws ec2 describe-vpcs --filters Name=isDefault,Values=true --query 'Vpcs[0].VpcId' --output text --region $REGION)
echo "VPC: $VPC_ID"

echo "== Creating security group =="
SG_ID=$(aws ec2 create-security-group --group-name "$SG_NAME" --description "Cloud Guard app - HTTP + SSH" --vpc-id "$VPC_ID" --region $REGION --query 'GroupId' --output text 2>/dev/null || \
  aws ec2 describe-security-groups --filters Name=group-name,Values=$SG_NAME --region $REGION --query 'SecurityGroups[0].GroupId' --output text)
echo "SG: $SG_ID"

aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --protocol tcp --port 80 --cidr 0.0.0.0/0 --region $REGION 2>/dev/null || true
aws ec2 authorize-security-group-ingress --group-id "$SG_ID" --protocol tcp --port 22 --cidr 0.0.0.0/0 --region $REGION 2>/dev/null || true

echo "== Getting latest Amazon Linux 2023 AMI =="
AMI_ID=$(aws ssm get-parameters --names /aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64 --region $REGION --query 'Parameters[0].Value' --output text 2>/dev/null || echo "")
if [ -z "$AMI_ID" ] || [ "$AMI_ID" == "None" ]; then
  echo "(SSM lookup not permitted for this IAM user, falling back to known AMI)"
  AMI_ID="ami-02b64aa047cb5edf5"
fi
echo "AMI: $AMI_ID"

echo "== Writing user-data (installs Docker, clones repo, builds, runs on port 80) =="
# Note: heredoc delimiter is unquoted on purpose so $REPO_URL expands below (no sed needed - avoids macOS/BSD sed -i incompatibility)
# Logs to /var/log/user-data.log so `cat /var/log/user-data.log` over SSH shows the full build output
# (console output truncates before the Docker build finishes).
cat > /tmp/user-data.sh << EOF
#!/bin/bash
exec > /var/log/user-data.log 2>&1
set -x

# Swap is REQUIRED here. Compiling aws-sdk-go-v2/service/ec2 peaks around 900MB-1.5GB,
# and EC2 instances ship with zero swap - a t3.small (1.9GB usable) OOM-killed the
# build mid-flight with no clean error, it just appeared to hang. 4GB swap + the
# -p 1 build flag in the Dockerfile keeps peak memory well inside the box.
dd if=/dev/zero of=/swapfile bs=1M count=4096
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
free -h

dnf install -y docker git
systemctl enable --now docker
cd /home/ec2-user
git clone $REPO_URL cloud-guard
cd cloud-guard
docker build -t cloud-guard .
docker run -d -p 80:8080 --name cloud-guard-app --restart unless-stopped -v cloudguard-data:/root/data cloud-guard
docker ps
echo "USER_DATA_SCRIPT_COMPLETE"
EOF

echo "== Ensuring key pair exists (so SSH is always possible) =="
# Without a key pair the ONLY way in is EC2 Instance Connect, which proved unreliable
# on this account - that left us with no way to inspect a stuck box. Always attach one.
if [ ! -f "$HOME/.ssh/$KEY_NAME.pem" ]; then
  aws ec2 delete-key-pair --key-name "$KEY_NAME" --region $REGION 2>/dev/null || true
  mkdir -p "$HOME/.ssh"
  aws ec2 create-key-pair --key-name "$KEY_NAME" --region $REGION \
    --query 'KeyMaterial' --output text > "$HOME/.ssh/$KEY_NAME.pem"
  chmod 400 "$HOME/.ssh/$KEY_NAME.pem"
  echo "Created new key at ~/.ssh/$KEY_NAME.pem"
else
  echo "Reusing existing key at ~/.ssh/$KEY_NAME.pem"
fi

echo "== Launching instance =="
# t3.small (2 GiB) + 4GB swap. t3.micro (1 GiB) and even t3.small without swap both
# OOM-killed the Go build silently. 20GB root volume so the swapfile + Docker layers fit.
INSTANCE_ID=$(aws ec2 run-instances \
  --image-id "$AMI_ID" \
  --instance-type t3.small \
  --key-name "$KEY_NAME" \
  --block-device-mappings 'DeviceName=/dev/xvda,Ebs={VolumeSize=20,VolumeType=gp3,DeleteOnTermination=true}' \
  --security-group-ids "$SG_ID" \
  --user-data file:///tmp/user-data.sh \
  --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=$INSTANCE_NAME}]" \
  --region $REGION \
  --query 'Instances[0].InstanceId' --output text)
echo "Instance: $INSTANCE_ID"

echo "== Waiting for it to be running =="
aws ec2 wait instance-running --instance-ids "$INSTANCE_ID" --region $REGION

PUBLIC_IP=$(aws ec2 describe-instances --instance-ids "$INSTANCE_ID" --region $REGION --query 'Reservations[0].Instances[0].PublicIpAddress' --output text)
echo ""
echo "=================================================="
echo "Instance is up: $INSTANCE_ID"
echo "Public IP: $PUBLIC_IP"
echo ""
echo "The Docker build (Go + CGO, large AWS SDK) takes ~5-10 min. Be patient."
echo ""
echo "  Test the app:   curl http://$PUBLIC_IP"
echo "  Watch the build: ssh -i ~/.ssh/$KEY_NAME.pem ec2-user@$PUBLIC_IP 'sudo tail -f /var/log/user-data.log'"
echo "  Check container: ssh -i ~/.ssh/$KEY_NAME.pem ec2-user@$PUBLIC_IP 'sudo docker ps -a'"
echo "=================================================="
