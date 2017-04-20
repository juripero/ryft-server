#!/bin/bash
set -e
set -o pipefail

echo "ETRYPOINT"
go env

if [ ! -d "${RYFTPATH}" ]; then
    go get -d -v github.com/getryft/ryft-server
fi

exec "$@"
