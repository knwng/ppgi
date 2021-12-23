#!/bin/sh

ADDRESS="http://192.168.31.147:8080/"
PULSAR_BASE_DIR="/Users/knwng/Downloads/apache-pulsar-2.8.1"

${PULSAR_BASE_DIR}/bin/pulsar-admin --admin-url ${ADDRESS} topics delete "persistent://public/default/host"

${PULSAR_BASE_DIR}/bin/pulsar-admin --admin-url ${ADDRESS} topics delete "persistent://public/default/client"
