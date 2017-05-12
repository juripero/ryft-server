SOURCE ?= `cd ../../../ && pwd`
VERSION ?= latest
REGISTRY ?= 172.16.34.3:5000

.PHONY: builder build app clear

builder:
	cd builder && \
	docker build --no-cache --rm -t=${REGISTRY}/ryft/builder:${VERSION} .

push_builder:
	docker push ${REGISTRY}/ryft/builder:${VERSION}


build:
	cd builder && \
	mkdir -p .tmp && \
	rsync -avz --exclude=".tmp" --exclude="debian" --exclude="docker" ${SOURCE} .tmp/ && \
	docker run -v ${SOURCE}:/go/src/github.com/getryft ${REGISTRY}/ryft/builder:${VERSION} && \
	rm -rf .tmp

app:
	cd app && \
	mkdir -p .tmp && \
	cp ${SOURCE}/ryft-server/ryft-server .tmp/ryft-server && \
	docker build -t=${REGISTRY}/ryft/app:${VERSION} . && \
	rm -rf .tmp

push_app:
	docker push ${REGISTRY}/ryft/app:${VERSION}

clear:
	rm -rfv app/.tmp
	rm -rfv builder/.tmp
	-docker rm `docker ps -qa`
	-docker rmi `docker images -f "dangling=true" -q`
