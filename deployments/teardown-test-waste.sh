#!/bin/bash
# Delete everything seed-test-waste.sh created.
#
# Deletion is driven strictly by the CloudGuardTest=true tag. Nothing here
# deletes by name, by age, or by "looks unused" - a teardown script that
# guesses is a script that eventually deletes something that mattered.

set -uo pipefail   # deliberately NOT -e: one failure must not abort the rest of the cleanup

REGION="${REGION:-us-east-1}"
TAG="CloudGuardTest"
FILTER="Name=tag:$TAG,Values=true"

echo "== Removing resources tagged $TAG=true in $REGION =="

# Snapshots must go before their source volumes.
echo ""
echo "-- Snapshots --"
for s in $(aws ec2 describe-snapshots --region "$REGION" --owner-ids self \
    --filters "$FILTER" --query 'Snapshots[].SnapshotId' --output text); do
  echo "   deleting $s"
  aws ec2 delete-snapshot --region "$REGION" --snapshot-id "$s"
done

# Attached volumes must be detached before they can be deleted.
echo ""
echo "-- Volumes --"
for v in $(aws ec2 describe-volumes --region "$REGION" \
    --filters "$FILTER" --query 'Volumes[].VolumeId' --output text); do
  state=$(aws ec2 describe-volumes --region "$REGION" --volume-ids "$v" \
    --query 'Volumes[0].State' --output text)
  if [ "$state" = "in-use" ]; then
    echo "   detaching $v"
    aws ec2 detach-volume --region "$REGION" --volume-id "$v" >/dev/null
    aws ec2 wait volume-available --region "$REGION" --volume-ids "$v"
  fi
  echo "   deleting $v"
  aws ec2 delete-volume --region "$REGION" --volume-id "$v"
done

echo ""
echo "-- Elastic IPs --"
for a in $(aws ec2 describe-addresses --region "$REGION" \
    --filters "$FILTER" --query 'Addresses[].AllocationId' --output text); do
  echo "   releasing $a"
  aws ec2 release-address --region "$REGION" --allocation-id "$a"
done

echo ""
echo "== Done. Verifying nothing is left =="
aws ec2 describe-volumes  --region "$REGION" --filters "$FILTER" --query 'Volumes[].VolumeId'    --output text
aws ec2 describe-addresses --region "$REGION" --filters "$FILTER" --query 'Addresses[].AllocationId' --output text
aws ec2 describe-snapshots --region "$REGION" --owner-ids self --filters "$FILTER" --query 'Snapshots[].SnapshotId' --output text
echo "(empty output above = clean)"
