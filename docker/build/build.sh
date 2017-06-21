#!/bin/sh -e
cd ${RYFTPATH}

echo "syncing vendor sources..."
govendor sync

# build Debian package
make debian

# build go binary file
# make build
