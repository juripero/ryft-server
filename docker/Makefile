HINT = ryft-server/docker

SOURCE ?= $(abspath $(dir ${CURDIR}))
VERSION ?= latest

all: debian

.PHONY: image debian build cli

# build ryft-server binary in Docker container
# the output can be found at ../ryft-server
# update vendor and build as $USER, not as root
build: image
	@echo "[${HINT}]: building ryft-server (as ${USER})..."
	docker run --rm \
		-v /etc/group:/etc/group:ro -v /etc/passwd:/etc/passwd:ro \
		--user $(shell id -u ${USER}):$(shell id -g ${USER}) \
		-v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/build:${VERSION} make all

# build Debian package
debian: build
	@echo "[${HINT}]: building Debian package (as root)..."
	docker run --rm \
		-v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/build:${VERSION} make -C debian all

# run /bin/bash in Docker container
# run as $USER, not as root
cli: image
	@echo "[${HINT}]: running CLI (as ${USER})..."
	docker run --rm -it \
		-v /etc/group:/etc/group:ro -v /etc/passwd:/etc/passwd:ro \
		--user $(shell id -u ${USER}):$(shell id -g ${USER}) \
		-v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/build:${VERSION} /bin/bash

# create image to build sources
image:
	@echo "[${HINT}]: building Docker image..."
	docker build -q -t=ryft/build:${VERSION} build
