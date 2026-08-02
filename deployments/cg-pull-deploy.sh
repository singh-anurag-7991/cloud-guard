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
# Bind to 127.0.0.1:8080, NOT 0.0.0.0:80.
#
# Caddy owns ports 80 and 443 on this box (it terminates TLS and auto-renews the
# Let's Encrypt cert) and reverse-proxies to 127.0.0.1:8080. Publishing the app
# on :80 fails with "address already in use", and publishing on 0.0.0.0:8080
# would expose a plaintext HTTP bypass around Caddy - session cookies are marked
# Secure, so that path would also silently break login.
docker run -d \
  -p 127.0.0.1:8080:8080 \
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
# Hit the app directly on 8080, not through Caddy on :80. Caddy only answers for
# the guardinfra.duckdns.org host header, so curling localhost:80 returns a 404
# from Caddy and tells you nothing about whether the app came up.
curl -s "http://127.0.0.1:8080/healthz?cb=$(date +%s)" || echo "(app still starting)"
echo ""
