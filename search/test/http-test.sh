#!/bin/bash
set -e # exit on any error
set -x # debugging

SERVER_ADDR="http://localhost:8765"
FORMAT_JSON='-H "Accept: application/json"'
FORMAT_MSGPACK='-H "Accept: application/msgpack"'
FORMAT="${FORMAT_JSON}"
#FORMAT="${FORMAT_MSGPACK}"

echo "server address: ${SERVER_ADDR}"
echo "accepted format: ${FORMAT}"

curl "${SERVER_ADDR}/search?query=555&files=/regression/*.txt&surrounding=10&fuzziness=0&format=raw&local=true"
