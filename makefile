GOBINDATA = $GOPATH/bin/go-bindata
ASSETS = bindata.go
BINARIES = ryft-server

all: $(ASSETS) build


$(GOBINDATA):
	go get -u github.com/jteeuwen/go-bindata/...


$(ASSETS): $(GOBINDATA)
	go-bindata -o bindata.go -prefix static/ static/...
	

.PHONY: build	
build:
	go build
	
.PHONY: install
install: $(ASSETS) 
	go install
	
	
clean: 
	rm -f $(ASSETS)
	rm -f $(BINARIES) 