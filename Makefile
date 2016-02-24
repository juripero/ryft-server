GOBINDATA = ${GOPATH}/bin/go-bindata
ASSETS = bindata.go
BINARIES = ryft-server

# disable ryftone search engine by default
ifeq (${GO_TAGS},)
  GO_TAGS=noryftone
endif

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
	go build -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"

.PHONY: install
install: $(ASSETS)
	go install -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"

.PHONY: debian
debian: install
	make -C debian package VERSION=${VERSION} GITHASH=${GITHASH}

clean:
	rm -f $(ASSETS)
	rm -f $(BINARIES)
