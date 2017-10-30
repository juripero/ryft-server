# do not use alpine because it uses incompatible libc!
FROM golang:1.8.5

RUN apt-get update && apt-get install -y git dpkg

# manage dependencies
RUN go get -u github.com/kardianos/govendor

# predifined source path
ENV RYFTPATH ${GOPATH}/src/github.com/getryft/ryft-server
RUN mkdir -p ${RYFTPATH}
WORKDIR ${RYFTPATH}

COPY build.sh /build.sh
RUN chmod +x /build.sh

CMD ["/build.sh"]
