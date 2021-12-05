#!/bin/sh

set -euxo pipefail

# deploy nebula-graph using docker-compose
# service will listen to 0.0.0.0:9669
# default username/password: root/nebula
git clone -b v2.6.0 https://github.com/vesoft-inc/nebula-docker-compose.git
cd nebula-docker-compose/
docker-compose up -d

# deploy nebula-studio using docker-compose
# service will listen to 0.0.0.0:7001
wget https://oss-cdn.nebula-graph.com.cn/nebula-graph-studio/nebula-graph-studio-v3.1.0.tar.gz
mkdir nebula-graph-studio-v3.1.0
tar -zxvf nebula-graph-studio-v3.1.0.tar.gz -C nebula-graph-studio-v3.1.0
cd nebula-graph-studio-v3.1.0
docker-compose up -d
