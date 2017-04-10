SOURCE = ${GOPATH}/src/github.com/getryft
VERSION ?= latest

.PHONY: builder build app clear

builder:
	cd builder && \
	docker build --no-cache --rm -t=ryft/builder:${VERSION} . && \
	docker tag ryft/builder:${VERSION} ryft/builder:latest

build:
	cd builder && \
	mkdir -p .tmp && \
	rsync -avz --exclude=".tmp" --exclude="debian" --exclude="docker" ${SOURCE} .tmp/ && \
	docker run -v ${SOURCE}:/go/src/github.com/getryft ryft/builder:${VERSION}
	rm -rf .tmp

app:
	cd app && \
	mkdir -p .tmp && \
	cp ${SOURCE}/ryft-server/ryft-server .tmp/ryft-server && \
	docker build -t=ryft/app:${VERSION} . && \
	rm -rf .tmp

clear:
	rm -rfv app/.tmp
	rm -rfv builder/.tmp
	docker rm `docker ps -qa`
	docker rmi `docker images -f "dangling=true" -q`