#!/bin/sh

set -euxo pipefail

docker run --name redis -p 6379:6379 -e ALLOW_EMPTY_PASSWORD=yes bitnami/redis:latest
