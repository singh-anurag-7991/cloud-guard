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
cat > /tmp/user-data.sh << EOF
#!/bin/bash
dnf install -y docker git
systemctl enable --now docker
cd /home/ec2-user
git clone $REPO_URL cloud-guard
cd cloud-guard
docker build -t cloud-guard .
docker run -d -p 80:8080 --name cloud-guard-app --restart unless-stopped -v cloudguard-data:/root/data cloud-guard
EOF

echo "== Launching instance =="
INSTANCE_ID=$(aws ec2 run-instances \
  --image-id "$AMI_ID" \
  --instance-type t3.micro \
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
echo "App will be live at http://$PUBLIC_IP within ~2 minutes (Docker build takes a bit)."
echo "Test with: curl http://$PUBLIC_IP"
echo "=================================================="
