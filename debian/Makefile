# support for VERSION AND GITHASH environment variables
HINT=ryft-server/debian

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
package-content: control
	@echo "[${HINT}]: saving package content..."
	@cp "${RYFT_SERVER_BIN}" "${DESTDIR}/usr/bin/"
	@cp "../ryft-server.conf" "${DESTDIR}/etc/ryft-server.conf"
	@cp "../ryft-users.yaml" "${DESTDIR}/etc/ryft-users.yaml"
	@echo "/etc/ryft-server.conf" > "${DESTDIR}/DEBIAN/conffiles"
	@echo "/etc/ryft-users.yaml" >> "${DESTDIR}/DEBIAN/conffiles"
	@rm "${DESTDIR}/var/log/ryft/.keep"
	@rm "${DESTDIR}/var/ryft/.keep"
	@echo "Installed-Size: $$(du -k -s ${DESTDIR} | awk '{print $$1}')" >> "${CONTROLFILE}"
# chown -R ryftuser.ryftuser "${DESTDIR}"

package: package-content
	@echo "[${HINT}]: building package..."
	@dpkg-deb --build "${DESTDIR}" "${DEBNAME}"
	@rm -rf "${DESTDIR}"

# initialize template
.PHONY: template
template:
	@echo "[${HINT}]: initializing package template..."
	@rm -rf ${DESTDIR}
	@mkdir -p ${DESTDIR}
	@cp -r template/* ${DESTDIR}/
	@mkdir -p ${DESTDIR}/usr/bin

# create DEBIAN control file
CONTROLFILE=${DESTDIR}/DEBIAN/control
.PHONY: control
control: template
	@echo "[${HINT}]: creating CONTROL file..."
	@echo "Package: ryft-server" > "${CONTROLFILE}"
	@echo "Version: ${VERSION}" >> "${CONTROLFILE}"
	@echo "Section: custom" >> "${CONTROLFILE}"
	@echo "Priority: optional" >> "${CONTROLFILE}"
	@echo "Architecture: amd64" >> "${CONTROLFILE}"
	@echo "Essential: no" >> "${CONTROLFILE}"
# @echo "Installed-Size: 1024" >> "${CONTROLFILE}"
	@echo "Maintainer: www.ryft.com" >> "${CONTROLFILE}"
	@echo "Description: ryft server" >> "${CONTROLFILE}"
