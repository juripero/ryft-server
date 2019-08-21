#!/bin/bash

# Script to verify setup for BlackLynx geo demo
GREEN='\033[1;32m'
RED='\033[1;31m'
NC='\033[0m'

SIDE=${1:-"client"}

displayLine() {
   DOTS="................................................................"
   echo -en "${1} ${DOTS:${#1}:60} "
}


if [ "${SIDE}" == "client" ]; then
    echo -e "Checking Client side setup:"
else
    echo -e "Checking Server side setup:"
	displayLine "...polygons directory location"
	if [ -d /home/ryftuser/.blacklynx/polygons ]; then
	    echo "${GREEN}OK${NC}"
    else
	    echo -e "${RED}ERROR${NC}"
	fi
fi
