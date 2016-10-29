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

.PHONY: $(ASSETS)
$(ASSETS): $(GOBINDATA)
	${GOBINDATA} -o bindata.go -prefix static/ static/...

.PHONY: update
update:
	go get -d -u -t -v ./...

.PHONY: build
build:
	go build -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"

.PHONY: install
install: $(ASSETS)
	go install -ldflags "-X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"

.PHONY: debian
debian: install
	make -C debian package VERSION=${VERSION} GITHASH=${GITHASH}


.PHONY: test_cover test-cover test
test_cover: test-cover
test-cover:
	@go test -tags "${GO_TAGS}" -cover ./search/ || true
# go test -tags "${GO_TAGS}" -cover ./search/ryftdec/ || true
# go test -tags "${GO_TAGS}" -cover ./search/ryfthttp/ || true
# go test -tags "${GO_TAGS}" -cover ./search/ryftmux/ || true
# go test -tags "${GO_TAGS}" -cover ./search/ryftone/ || true
# go test -tags "${GO_TAGS}" -cover ./search/ryftprim/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/ || true
# go test -tags "${GO_TAGS}" -cover ./search/utils/catalog/ || true

# go test -tags "${GO_TAGS}" -cover ./rest/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/json/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/msgpack.v1/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/msgpack.v2/ || true
# go test -tags "${GO_TAGS}" -cover ./rest/format/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/raw/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/null/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/utf8/ || true
# go test -tags "${GO_TAGS}" -cover ./rest/format/json/ || true
# go test -tags "${GO_TAGS}" -cover ./rest/format/xml/ || true

	@go test -tags "${GO_TAGS}" -cover ./middleware/auth/ || true
	@go test -tags "${GO_TAGS}" -cover ./middleware/cors/ || true
	@go test -tags "${GO_TAGS}" -cover ./middleware/gzip/ || true

test:
	go test -v ./...

clean:
	rm -f $(ASSETS)
	rm -f $(BINARIES)
