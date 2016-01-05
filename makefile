#!/bin/bash

SWAGGER_FILE="swagger.json"
INDEX_FILE="index.html"

function make {
  generateSwaggerDoc
  generateAssets
  clean
  buildRyftRest
}

function generateSwaggerDoc {
  if [ -f $GOPATH"/swagger" ]; then
  echo "Getting go-swagger"
  eval go get -u github.com/go-swagger/go-swagger/cmd/swagger
  fi
  echo "Swagger spec generating started"
  eval swagger generate spec -o $SWAGGER_FILE
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

function clean {
  eval rm $SWAGGER_FILE
}

make
