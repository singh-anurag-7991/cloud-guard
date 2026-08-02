#!/bin/bash
# Create a small, real, deliberately-wasteful set of AWS resources so Cloud
# Guard's cost rules have something genuine to find.
#
# WHY REAL RESOURCES AND NOT FAKE DATABASE ROWS:
# Seeding the SQLite database directly would prove nothing. The whole question
# we need answered is "does the scanner actually see waste in a live AWS
# account and price it correctly", and only real resources answer that.
#
# THIS COSTS REAL MONEY - about $0.15/day (~$4.60/month) if left running.
# It is small on purpose, but it is not free. Run the teardown script when done:
#     ./deployments/teardown-test-waste.sh
#
# Everything created here is tagged CloudGuardTest=true, and the teardown script
# deletes strictly by that tag, so it can never touch your real resources.

set -euo pipefail

REGION="${REGION:-us-east-1}"
TAG="CloudGuardTest"

echo "== Creating test waste in $REGION =="
echo "   (all resources tagged $TAG=true)"
echo ""

AZ=$(aws ec2 describe-availability-zones --region "$REGION" \
  --query 'AvailabilityZones[0].ZoneName' --output text)
echo "AZ: $AZ"

# ---------------------------------------------------------------------------
# 1. Unattached gp2 volume -> EBSUnattachedRule
#    8 GiB gp2 = 8 x $0.10 = $0.80/mo
#    This is the single most common form of real cloud waste.
# ---------------------------------------------------------------------------
echo ""
echo "-- 1/4 Unattached EBS volume (gp2, 8 GiB) --"
VOL_UNATTACHED=$(aws ec2 create-volume \
  --region "$REGION" --availability-zone "$AZ" \
  --size 8 --volume-type gp2 \
  --tag-specifications "ResourceType=volume,Tags=[{Key=$TAG,Value=true},{Key=Name,Value=cg-test-unattached}]" \
  --query 'VolumeId' --output text)
echo "   Created $VOL_UNATTACHED  (expect: \$0.80/mo saving, HIGH confidence)"

# ---------------------------------------------------------------------------
# 2. Unassociated Elastic IP -> ElasticIPUnusedRule
#    $0.005/hour x 730 = $3.65/mo. AWS charges precisely because it is idle.
# ---------------------------------------------------------------------------
echo ""
echo "-- 2/4 Unassociated Elastic IP --"
EIP_ALLOC=$(aws ec2 allocate-address \
  --region "$REGION" --domain vpc \
  --tag-specifications "ResourceType=elastic-ip,Tags=[{Key=$TAG,Value=true},{Key=Name,Value=cg-test-idle-eip}]" \
  --query 'AllocationId' --output text)
echo "   Created $EIP_ALLOC  (expect: \$3.65/mo saving, HIGH confidence)"

# ---------------------------------------------------------------------------
# 3. Snapshot -> SnapshotStaleRule
#    AWS gives no way to back-date a snapshot, so this one only fires if the
#    app runs with CLOUDGUARD_SNAPSHOT_STALE_DAYS=0. That env var is set on the
#    container by the deploy command in the README for exactly this reason.
# ---------------------------------------------------------------------------
echo ""
echo "-- 3/4 Snapshot (for the stale-snapshot rule) --"
aws ec2 wait volume-available --region "$REGION" --volume-ids "$VOL_UNATTACHED"
SNAP=$(aws ec2 create-snapshot \
  --region "$REGION" --volume-id "$VOL_UNATTACHED" \
  --description "Cloud Guard test snapshot - safe to delete" \
  --tag-specifications "ResourceType=snapshot,Tags=[{Key=$TAG,Value=true},{Key=Name,Value=cg-test-snapshot}]" \
  --query 'SnapshotId' --output text)
echo "   Created $SNAP  (expect: \$0.40/mo saving - ONLY with CLOUDGUARD_SNAPSHOT_STALE_DAYS=0)"

# ---------------------------------------------------------------------------
# 4. Attached gp2 volume -> EBSGp2ToGp3Rule
#    Attached because the rule deliberately skips unattached gp2 volumes (those
#    are already covered by rule 1, and double-counting would inflate the total).
#    8 GiB x ($0.10 - $0.08) = $0.16/mo
# ---------------------------------------------------------------------------
echo ""
echo "-- 4/4 Attached gp2 volume (for the gp2->gp3 rule) --"
INSTANCE_ID=$(aws ec2 describe-instances --region "$REGION" \
  --filters "Name=tag:Name,Values=cloud-guard-prod" "Name=instance-state-name,Values=running" \
  --query 'Reservations[0].Instances[0].InstanceId' --output text 2>/dev/null || echo "None")

if [ "$INSTANCE_ID" = "None" ] || [ -z "$INSTANCE_ID" ]; then
  echo "   SKIPPED: no running cloud-guard-prod instance found to attach to."
  VOL_GP2="none"
else
  INSTANCE_AZ=$(aws ec2 describe-instances --region "$REGION" --instance-ids "$INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].Placement.AvailabilityZone' --output text)
  VOL_GP2=$(aws ec2 create-volume \
    --region "$REGION" --availability-zone "$INSTANCE_AZ" \
    --size 8 --volume-type gp2 \
    --tag-specifications "ResourceType=volume,Tags=[{Key=$TAG,Value=true},{Key=Name,Value=cg-test-gp2}]" \
    --query 'VolumeId' --output text)
  aws ec2 wait volume-available --region "$REGION" --volume-ids "$VOL_GP2"
  # /dev/sdf is a spare device name. The volume is never formatted or mounted,
  # so nothing on the instance changes - it just has to be "in-use" for the rule.
  aws ec2 attach-volume --region "$REGION" \
    --volume-id "$VOL_GP2" --instance-id "$INSTANCE_ID" --device /dev/sdf >/dev/null
  echo "   Created $VOL_GP2 and attached to $INSTANCE_ID  (expect: \$0.16/mo saving)"
fi

cat <<SUMMARY

==========================================================
Test waste created. Expected dashboard total: ~\$4.61/mo
  \$0.80  unattached 8 GiB gp2 volume   ($VOL_UNATTACHED)
  \$3.65  unassociated Elastic IP       ($EIP_ALLOC)
  \$0.40  snapshot (needs STALE_DAYS=0) ($SNAP)
  \$0.16  gp2 -> gp3 migration          ($VOL_GP2)

Next:
  1. Trigger a scan from the dashboard
  2. Check each figure above appears against the right resource ID
  3. TEAR IT DOWN:  ./deployments/teardown-test-waste.sh

This is billing you roughly \$0.15/day until you do step 3.
==========================================================
SUMMARY
