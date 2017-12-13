# build caggs tool
#
# make caggs - builds tool on current platform
# make caggs PLATFORM=arm - buils for ARM (using cross-compilation)
# make clean - cleans all build artefacts

# select cross-compiler
ifeq (${PLATFORM},arm)
  CROSS=arm-linux-gnueabihf-
endif

ifeq (${VERSION},)
  VERSION=$(shell git describe --tags)
endif
ifeq (${GITHASH},)
  GITHASH=$(shell git log -1 --format='%H')
endif
ifeq (${BUILDTIME},)
  BUILDTIME=$(shell date -u '+%Y-%m-%d %H:%M:%S')
endif

DEFINES=-D_GNU_SOURCE
DEFINES+=-DCAGGS_VERSION='"${VERSION}"'
DEFINES+=-DCAGGS_GITHASH='"${GITHASH}"'
DEFINES+=-DCAGGS_BUILDTIME='"${BUILDTIME}"'

.PHONY: caggs caggs-stripped clean
all: caggs-stripped

# build caggs executable
caggs: main.c misc.c misc.h conf.c conf.h json.c json.h
	${CROSS}gcc -std=c99 ${DEFINES} -Wall -O3 -pthread -o caggs main.c misc.c conf.c json.c

# strip executable
caggs-stripped: caggs
	${CROSS}strip caggs

clean:
	rm -f *.o caggs
