#!/bin/sh

set -euxo pipefail

git clone -b v2.6.0 https://github.com/vesoft-inc/nebula-importer.git
(cd nebula-importer && cd nebula-importer)

./nebula-importer/nebula-importer -config ./load_graph.yaml
