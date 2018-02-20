TEMPLATE = app
CONFIG += console
CONFIG -= app_bundle
CONFIG -= qt

QMAKE_CFLAGS += -std=c++0x -pthread
LIBS += -pthread

DEFINES += _GNU_SOURCE

SOURCES += main.c
HEADERS += conf.h
SOURCES += conf.c
HEADERS += misc.h
SOURCES += misc.c
HEADERS += json.h
SOURCES += json.c
HEADERS += stat.h
SOURCES += stat.c

OTHER_FILES += README.md
OTHER_FILES += Makefile
