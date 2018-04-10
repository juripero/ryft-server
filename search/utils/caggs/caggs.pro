TEMPLATE = app
CONFIG += console
CONFIG -= app_bundle
CONFIG -= qt

QMAKE_CFLAGS += -std=c99 -pthread
LIBS += -pthread

DEFINES += _GNU_SOURCE
DEFINES += NO_TESTS

SOURCES += main.c
HEADERS += conf.h
SOURCES += conf.c
HEADERS += misc.h
SOURCES += misc.c
HEADERS += json.h
SOURCES += json.c
HEADERS += stat.h
SOURCES += stat.c
HEADERS += proc.h
SOURCES += proc.c
SOURCES += test.c

OTHER_FILES += README.md
OTHER_FILES += Makefile
