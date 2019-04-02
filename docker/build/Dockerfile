# do not use alpine because it uses incompatible libc!

#build stopped working when jessie stopped LTS update support
#FROM golang:1.8.5

#BUILD_HOMEDIR workaround addresses golang:1.12.1 issue where non-root user tries to create ~/.cache directory 
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
