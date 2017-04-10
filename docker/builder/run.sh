#!/bin/bash

echo "RUNNER"
cd ${RYFTPATH}
# get deps
govendor sync
# install into gopath
make install 
# build go binary file 
make build