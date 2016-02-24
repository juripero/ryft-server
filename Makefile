GOBINDATA = ${GOPATH}/bin/go-bindata
ASSETS = bindata.go
BINARIES = ryft-server

all: $(ASSETS) build

ifeq (${VERSION},)
  VERSION=$(shell git describe --tags)
endif
ifeq (${GITHASH},)
  GITHASH=$(shell git log -1 --format='%H')
endif

.PHONY: version
version:
	@echo "Version: ${VERSION}"
	@echo "githash: ${GITHASH}"

$(GOBINDATA):
	go get -u github.com/jteeuwen/go-bindata/...

$(ASSETS): $(GOBINDATA)
	${GOBINDATA} -o bindata.go -prefix static/ static/...

.PHONY: build
build:
	go build -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GITHASH}"

.PHONY: install
install: $(ASSETS)
	go install -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GITHASH}"

.PHONY: debian
debian: install
	make -C debian package VERSION=${VERSION} GITHASH=${GITHASH}

clean:
	rm -f $(ASSETS)
	rm -f $(BINARIES)
