# Docker file to test ryft-server building.
# Get's sources from current directory (actually ..)
FROM golang:1.8.1

RUN apt-get update -y && apt-get install -y \
     git \
     dpkg \
     tar \
     less \
     vim \
     jq \
     curl

ENV RYFTPATH ${GOPATH}/src/github.com/getryft/ryft-server

# manage dependencies
RUN go get -u github.com/kardianos/govendor

# copy entrypoint and runner scripts
COPY entrypoint.sh /entrypoint.sh
COPY run.sh /run.sh
RUN chmod +x /entrypoint.sh && \
    chmod +x /run.sh

# predifined source path
VOLUME ${GOPATH}/src/github.com/
WORKDIR ${RYFTPATH}

ENTRYPOINT ["/bin/bash", "/entrypoint.sh"]

CMD ["/bin/bash", "/run.sh"]
