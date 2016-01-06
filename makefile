#!/bin/bash

SWAGGER_FILE="swagger.json"
INDEX_FILE="index.html"

function make {
  generateAssets
  buildRyftRest
}


function generateAssets {
  if [ -f $GOPATH"/go-bindata" ]; then
  echo "Getting go-bindata"
  eval go get -u github.com/jteeuwen/go-bindata/...
  fi
  echo "Creating Assets"
  eval go-bindata -o bindata.go $INDEX_FILE $SWAGGER_FILE

}

function buildRyftRest {
  echo "Building ryft-server"
  eval go install
  echo "Starting ryft-rest"
  eval ryft-server
}


make
