#!/bin/bash

SWAGGER_FILE="swagger.json"
INDEX_FILE="index.html"

echo "Swagger spec generating started"
eval swagger generate spec -o $SWAGGER_FILE
echo "Getting go-bindata"
eval go get -u github.com/jteeuwen/go-bindata/...
echo "Creating Assets"
eval go-bindata -o bindata.go $INDEX_FILE $SWAGGER_FILE
echo "Building ryft-server"
eval go install
eval rm $SWAGGER_FILE
echo "Finished"
