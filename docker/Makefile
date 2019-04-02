HINT = ryft-server/docker

SOURCE ?= $(abspath $(dir ${CURDIR}))
BUILD_IMAGE_TAG ?= $(shell whoami) 

all: debian

.PHONY: image debian build cli debtorpm

# build ryft-server binary in Docker container
# the output can be found at ../ryft-server
# update vendor and build as $USER, not as root
build: image 
	@echo "[${HINT}]: building ryft-server (as ${USER})..."
	docker run --rm \
		--user $(shell id -u ${USER}):$(shell id -g ${USER}) \
		-v /etc/passwd:/etc/passwd:ro -v /etc/group:/etc/group:ro \
		-v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/server/build:${BUILD_IMAGE_TAG} make all

# build Debian package
debian: build
	@echo "[${HINT}]: building Debian package (as ${USER})..."
	docker run --rm \
	        --user $(shell id -u ${USER}):$(shell id -g ${USER}) \
		-v /etc/passwd:/etc/passwd:ro -v /etc/group:/etc/group:ro \
		-v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/server/build:${BUILD_IMAGE_TAG} make -C debian all

rpm: debian
	@echo "[${HINT}]: building RPM package (as root)..."
	@echo "SOURCE=${SOURCE}"
	docker run --rm \
		-v ${SOURCE}/docker:/opt/work \
		-v ${SOURCE}/debian:/opt/debian \
		debtorpm:${BUILD_IMAGE_TAG} /opt/work/build_rpm.sh /opt/debian/*.deb

# run /bin/bash in Docker container
# run as $USER, not as root
cli: image
	@echo "[${HINT}]: running CLI (as ${USER})..."
	docker run --rm -it \
		--user $(shell id -u ${USER}):$(shell id -g ${USER}) \
		-v /etc/passwd:/etc/passwd:ro -v /etc/group:/etc/group:ro \
		-v ${SOURCE}:/go/src/github.com/getryft/ryft-server \
		ryft/server/build:${BUILD_IMAGE_TAG} /bin/bash

# create image to build sources
image: debtorpm
	@echo "[${HINT}]: building Docker image..."
	docker build -q -t=ryft/server/build:${BUILD_IMAGE_TAG} --build-arg BUILD_HOMEDIR=$(shell grep `whoami` /etc/passwd|cut -d':' -f6) build

debtorpm: 
	@echo "[${HINT}]: building debtorpm Docker image..."
	docker build -q -t=debtorpm:${BUILD_IMAGE_TAG} debtorpm 
