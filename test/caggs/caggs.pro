TEMPLATE = app
CONFIG += console
CONFIG -= app_bundle
CONFIG -= qt

SOURCES += main.c
HEADERS += conf.h
SOURCES += conf.c
HEADERS += misc.h
SOURCES += misc.c

OTHER_FILES += README.md
OTHER_FILES += Makefile
