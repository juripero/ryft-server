SOURCE ?= $(dir ${CURDIR})
VERSION ?= latest

all: build

.PHONY: build-image build

# build Debian package in Docker container
# the output will be placed in ../debian/
build: build-image
	docker run --rm -v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/build:${VERSION}

# create image to build sources
build-image:
	docker build -q -t=ryft/build:${VERSION} build
