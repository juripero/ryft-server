FROM golang:1.8-alpine

RUN apk add --update git build-base dpkg tar xz

# manage dependencies
RUN go get -u github.com/kardianos/govendor

# predifined source path
ENV RYFTPATH ${GOPATH}/src/github.com/getryft/ryft-server
RUN mkdir -p ${RYFTPATH}
WORKDIR ${RYFTPATH}

COPY build.sh /build.sh
RUN chmod +x /build.sh

ENTRYPOINT ["/build.sh"]
