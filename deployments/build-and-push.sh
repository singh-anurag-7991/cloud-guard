#!/bin/bash
# Build the Cloud Guard image on your Mac and push it to Docker Hub.
#
# Why: building on the t3.small takes ~22 minutes because the Go compiler needs
# ~1.5GB while the box only has 1.9GB, so it thrashes on swap (observed: 22 min
# wall clock for 3 min of actual CPU). Your Mac has the RAM and cores to do the
# same build in 1-2 minutes. The server then just pulls a finished image.
#
# Usage:  DOCKER_USER=yourdockerhubusername ./deployments/build-and-push.sh
set -euo pipefail

: "${DOCKER_USER:?Set DOCKER_USER to your Docker Hub username}"
IMAGE="${DOCKER_USER}/cloud-guard"
SHA="$(git rev-parse --short HEAD)"

echo "== Building ${IMAGE}:${SHA} =="
# linux/amd64 is required: your Mac is arm64, the EC2 instance is x86_64.
docker build \
  --platform linux/amd64 \
  --build-arg GIT_SHA="${SHA}" \
  -t "${IMAGE}:${SHA}" \
  -t "${IMAGE}:latest" \
  .

echo "== Pushing =="
docker push "${IMAGE}:${SHA}"
docker push "${IMAGE}:latest"

echo ""
echo "Pushed ${IMAGE}:${SHA}"
echo "Now deploy with:"
echo "  ssh -o StrictHostKeyChecking=no -i ~/.ssh/cloud-guard-key.pem ec2-user@44.216.212.91 \\"
echo "    'sudo IMAGE=${IMAGE}:${SHA} /usr/local/bin/cg-pull-deploy.sh'"
