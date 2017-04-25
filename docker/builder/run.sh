#!/bin/bash

echo "RUNNER"
cd ${RYFTPATH}

# get deps
govendor sync

# build Debian package
make debian

# build go binary file
make build