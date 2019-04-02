# do not use alpine because it uses incompatible libc!
#FROM golang:1.8.5
FROM golang:1.12.1

ARG BUILD_HOMEDIR=/var/lib/jenkins

RUN apt-get update && apt-get install -y git dpkg

# predefined source path
ENV RYFTPATH ${GOPATH}/src/github.com/getryft/ryft-server
RUN mkdir -p ${RYFTPATH}
WORKDIR ${RYFTPATH}

# main dependencies
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN go get -u github.com/kardianos/govendor

RUN mkdir -p ${BUILD_HOMEDIR}
RUN chmod 777 ${BUILD_HOMEDIR}

ENTRYPOINT [ ]
CMD [ "make", "debian" ]
