GOBINDATA = ${GOPATH}/bin/go-bindata
ASSETS = bindata.go
BINARIES = ryft-server
HINT=ryft-server
APP_VERSION ?= latest

all: $(ASSETS) build version

ifeq (${VERSION},)
  VERSION=$(shell git describe --tags)
endif
ifeq (${GITHASH},)
  GITHASH=$(shell git log -1 --format='%H')
endif

.PHONY: version
version:
	@echo "Version: ${VERSION}"
	@echo "GitHash: ${GITHASH}"

$(GOBINDATA):
	@go get -u github.com/jteeuwen/go-bindata/...

.PHONY: $(ASSETS)
$(ASSETS): $(GOBINDATA)
	@echo "[${HINT}]: updating bindata..."
	@${GOBINDATA} -o bindata.go -prefix static/ static/...

.PHONY: update
update:
	go get -d -u -t -v ./...

.PHONY: build
build:
	@echo "[${HINT}]: building ryft-server..."
	@go build -ldflags "-s -w -X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"

.PHONY: install
install: $(ASSETS)
	@echo "[${HINT}]: installing ryft-server..."
	@go install -ldflags "-s -w -X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"

.PHONY: debian
debian: install
	@make -C debian package VERSION=${VERSION} GITHASH=${GITHASH}

# build Debian package using Docker container
.PHONY: docker-build docker_build
docker_build: docker-build
docker-build:
	@make -C docker build


.PHONY: test_cover test-cover test
test_cover: test-cover
test-cover:
	@go test -tags "${GO_TAGS}" -cover ./search/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryftdec/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryfthttp/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryftmux/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryftprim/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/catalog/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/query || true

	@go test -tags "${GO_TAGS}" -cover ./rest/codec/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/json/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/msgpack.v1/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/msgpack.v2/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/raw/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/null/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/utf8/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/json/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/xml/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/ || true

# @go test -tags "${GO_TAGS}" -cover ./middleware/auth/ || true
# @go test -tags "${GO_TAGS}" -cover ./middleware/cors/ || true
# @go test -tags "${GO_TAGS}" -cover ./middleware/gzip/ || true

test:
	go test -tags "${GO_TAGS}" ./...

clean:
	rm -f $(ASSETS)
	rm -f $(BINARIES)
.PHONY: build_container
build_container:
	if [ ! -d ./ryft-docker ]; then git clone git@github.com:getryft/ryft-docker.git; fi
	@make -C ./ryft-docker/ryft-server-cluster SOURCE_PATH=../../../../ build
	@make -C ./ryft-docker/ryft-server-cluster VERSION=${APP_VERSION} app

.PHONY: integration_test
# integration_test: build_container
integration_test:
	if [ ! -d ./ryft-integration-test ]; then git clone git@github.com:getryft/ryft-integration-test.git; fi
	@make -C ./ryft-integration-test APP_VERSION=${APP_VERSION} TEST_TAGS="not ryftx and not compound" all
