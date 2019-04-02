# do not use alpine because it uses incompatible libc!
#FROM golang:1.8.5
FROM golang:1.12.1

RUN apt-get update && apt-get install -y git dpkg

# predefined source path
ENV RYFTPATH ${GOPATH}/src/github.com/getryft/ryft-server
RUN mkdir -p ${RYFTPATH}
WORKDIR ${RYFTPATH}

# main dependencies
RUN go get -u github.com/jteeuwen/go-bindata/...
RUN go get -u github.com/kardianos/govendor

RUN mkdir /home/ryftuser
RUN chmod 777 /home/ryftuser

USER ryftuser

ENTRYPOINT [ ]
CMD [ "make", "debian" ]
