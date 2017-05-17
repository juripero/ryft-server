#!/bin/bash

if [[ -n "${CONSUL_HTTP_ADDR}" ]]; then
    curl -X PUT \
        --data '{"ID":"'${HOSTNAME}'", "Name":"ryft-rest-api","Tags":[],"Address":"'${HOSTNAME}'","Port":'${APP_PORT}'}' \
        "http://${CONSUL_HTTP_ADDR}/v1/agent/service/register"
fi

./ryft-server --config=/etc/ryft-server.conf --address :8765 --debug
