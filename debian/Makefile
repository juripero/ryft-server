# support for VERSION AND GITHASH environment variables

all: package

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

RYFT_SERVER_BIN=${GOPATH}/bin/ryft-server
DEBNAME=ryft-server-${VERSION}_amd64.deb
DESTDIR=.build

.PHONY: package
package: control
	cp "${RYFT_SERVER_BIN}" "${DESTDIR}/usr/bin/"
# TODO: copy server configuration	?
	dpkg --build "${DESTDIR}"
	mv -f "${DESTDIR}.deb" "${DEBNAME}"
	rm -rf "${DESTDIR}"

# initialize template
.PHONY: template
template:
	rm -rf ${DESTDIR}
	mkdir -p ${DESTDIR}
	cp -r template/* ${DESTDIR}/
	mkdir -p ${DESTDIR}/usr/bin

# create DEBIAN control file
CONTROLFILE=${DESTDIR}/DEBIAN/control
.PHONY: control
control: template
	echo "Package: ryft-server" > "${CONTROLFILE}"
	echo "Version: ${VERSION}" >> "${CONTROLFILE}"
	echo "Section: custom" >> "${CONTROLFILE}"
	echo "Priority: optional" >> "${CONTROLFILE}"
	echo "Architecture: amd64" >> "${CONTROLFILE}"
	echo "Essential: no" >> "${CONTROLFILE}"
	echo "Installed-Size: 1024" >> "${CONTROLFILE}"
	echo "Maintainer: www.ryft.com" >> "${CONTROLFILE}"
	echo "Description: ryft server" >> "${CONTROLFILE}"