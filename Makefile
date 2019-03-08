HINT = ryft-server
APP_VERSION ?= latest
RYFT_DOCKER_BRANCH ?= master
RYFT_INTEGRATION_TEST ?= develop

all: build caggs version

ifeq (${VERSION},)
  VERSION=$(shell git describe --tags|tr '-' '-')
endif
ifeq (${GITHASH},)
  GITHASH=$(shell git log -1 --format='%H')
endif

.PHONY: version version-q
version:
	@echo "Version: ${VERSION}"
	@echo "GitHash: ${GITHASH}"
version-q:
	@echo "${VERSION}"

# update binary data
.PHONY: update_bindata
update_bindata: ${GOPATH}/bin/go-bindata
	@echo "[${HINT}]: updating bindata..."
	@${GOPATH}/bin/go-bindata -o bindata.go -prefix static/ static/...
${GOPATH}/bin/go-bindata:
	@go get -u github.com/jteeuwen/go-bindata/...

# update vendor sources
.PHONY: update_vendor fetch_vendor
update_vendor: ${GOPATH}/bin/govendor
	@echo "[${HINT}]: updating vendor..."
	@${GOPATH}/bin/govendor sync
fetch_vendor: ${GOPATH}/bin/govendor
# looking for new version of dependencies
	@echo "[${HINT}]: fetching vendor..."
	@${GOPATH}/bin/govendor fetch -v +vendor
${GOPATH}/bin/govendor:
	@go get -u github.com/kardianos/govendor


# build or install
.PHONY: build install
build: update_vendor update_bindata
	@echo "[${HINT}]: building ryft-server..."
	@go build -ldflags "-s -w -X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"
install: update_vendor update_bindata
	@echo "[${HINT}]: installing ryft-server..."
	@go install -ldflags "-s -w -X main.Version=${VERSION} -X main.GitHash=${GITHASH}" -tags "${GO_TAGS}"


# build caggs tool
caggs:
	@echo "[${HINT}]: building ryft-server-aggs tool..."
	@make -C search/utils/caggs all
	@mv ./search/utils/caggs/caggs ./ryft-server-aggs


# build Debian package
.PHONY: debian docker-build docker_build
debian: build
	@make -C debian package VERSION=${VERSION} GITHASH=${GITHASH}
docker_build: docker-build
docker-build:
	@make -C docker rpm


# run coverage test
.PHONY: test_cover test-cover test
test_cover: test-cover
test-cover:
	@go test -tags "${GO_TAGS}" -cover ./search/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryftdec/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryfthttp/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryftmux/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/ryftprim/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/aggs/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/catalog/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/query/ || true
	@go test -tags "${GO_TAGS}" -cover ./search/utils/view/ || true

	@go test -tags "${GO_TAGS}" -cover ./rest/codec/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/json/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/msgpack.v1/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/codec/msgpack.v2/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/raw/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/null/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/utf8/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/json/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/xml/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/csv/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/format/ || true
	@go test -tags "${GO_TAGS}" -cover ./rest/ || true

# @go test -tags "${GO_TAGS}" -cover ./middleware/auth/ || true
# @go test -tags "${GO_TAGS}" -cover ./middleware/cors/ || true
# @go test -tags "${GO_TAGS}" -cover ./middleware/gzip/ || true

test:
	go test -tags "${GO_TAGS}" ./...

clean:
	@rm -f bindata.go
	@rm -f ryft-server


# clone and pull ryft-docker remote repository
.PHONY: clone_ryft_docker pull_ryft_docker
clone_ryft_docker:
	[ -d ryft-docker/.git ] || git clone git@github.com:getryft/ryft-docker.git
pull_ryft_docker: clone_ryft_docker
	cd ryft-docker && git fetch && git checkout ${RYFT_DOCKER_BRANCH} && git pull --ff-only

# clone and pull ryft-integration-test remote repostiory
.PHONY: clone_ryft_integration_test pull_ryft_integration_test
clone_ryft_integration_test:
	[ -d ryft-integration-test/.git ] || git clone git@github.com:getryft/ryft-integration-test.git
pull_ryft_integration_test: clone_ryft_integration_test
	cd ryft-integration-test && git fetch && git checkout ${RYFT_INTEGRATION_TEST} && git pull --ff-only


# build Docker containers with test environment
.PHONY: build_container
build_container: pull_ryft_docker
	@make -C ./ryft-docker/ryft-server-cluster SOURCE_PATH=${CURDIR}/ build
	@make -C ./ryft-docker/ryft-server-cluster APP_VERSION=${APP_VERSION} app

# run integration tests
.PHONY: integration_test
integration_test: pull_ryft_integration_test
	@make -C ./ryft-integration-test APP_VERSION=${APP_VERSION} TEST_TAGS="not ryftx and not compound and not backend_selection" all

# run unit tests
.PHONY: unit_test
unit_test: pull_ryft_docker
	@make -C ./ryft-docker/ryft-server-cluster SOURCE_PATH=${CURDIR} unit_test

.PHONY: cli
cli: pull_ryft_docker
	@make -C ./ryft-docker/ryft-server-cluster SOURCE_PATH=${CURDIR} cli
