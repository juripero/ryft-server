#!/bin/bash

echo "RUNNER"
cd ${RYFTPATH}

# get deps
govendor sync

# build go binary file
make build

# build Debian package
make debian
