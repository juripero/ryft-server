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

  eval go get -u github.com/go-swagger/go-swagger/cmd/swagger
  echo "Swagger spec generating started"
  eval swagger generate spec -o $SWAGGER_FILE
}

function generateAssets {
  echo "Getting go-bindata"
  eval go get -u github.com/jteeuwen/go-bindata/...
  echo "Creating Assets"
  eval go-bindata -o bindata.go $INDEX_FILE $SWAGGER_FILE

}

function buildRyftRest {
  echo "Building ryft-server"
  eval go install
  echo "Finished"
  eval ryft-server
}

function clean {
  eval rm $SWAGGER_FILE
}

make
