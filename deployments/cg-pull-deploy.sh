#!/bin/bash
# Pull a pre-built image from the registry and swap the running container.
#
# Lives in the repo (not just on the server) so the deploy procedure is version
# controlled. Copy it to the box with:
#   scp -i ~/.ssh/cloud-guard-key.pem deployments/cg-pull-deploy.sh ec2-user@44.216.212.91:/tmp/
#   ssh ... 'sudo mv /tmp/cg-pull-deploy.sh /usr/local/bin/ && sudo chmod +x /usr/local/bin/cg-pull-deploy.sh'
#
# Usage on the server:
#   sudo IMAGE=anurag7979doc/cloud-guard:SHA /usr/local/bin/cg-pull-deploy.sh

set -euo pipefail

: "${IMAGE:?Set IMAGE to the registry tag to deploy}"
CONTAINER="cloud-guard-app"

echo "== Pulling $IMAGE =="
docker pull "$IMAGE"

echo "== Stopping old container =="
docker rm -f "$CONTAINER" 2>/dev/null || true

echo "== Starting new container =="
# -v cloudguard-data:/root/data is what keeps users, sessions and findings alive
# across deploys. The DB path default was once /root/cloudguard.db, outside this
# volume, which silently wiped every account on each deploy.
#
# CLOUDGUARD_SNAPSHOT_STALE_DAYS: AWS provides no way to create a back-dated
# snapshot, so the 90-day default cannot be exercised against a freshly seeded
# test account. Set to 0 while validating; remove it before real customers.
docker run -d \
  -p 80:8080 \
  --name "$CONTAINER" \
  --restart unless-stopped \
  -v cloudguard-data:/root/data \
  -e CLOUDGUARD_SNAPSHOT_STALE_DAYS="${CLOUDGUARD_SNAPSHOT_STALE_DAYS:-90}" \
  ${CLOUDGUARD_REGIONS:+-e CLOUDGUARD_REGIONS="$CLOUDGUARD_REGIONS"} \
  ${CLOUDGUARD_EXTERNAL_ID:+-e CLOUDGUARD_EXTERNAL_ID="$CLOUDGUARD_EXTERNAL_ID"} \
  "$IMAGE"

sleep 2
docker ps --filter "name=$CONTAINER" --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'

echo ""
echo "== Version check =="
curl -s "http://localhost/healthz?cb=$(date +%s)" || echo "(app still starting)"
echo ""
