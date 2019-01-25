GO_PATH_PREFIX = GOPATH=$(shell pwd)
GO_CMD = $(GO_PATH_PREFIX) go
GO_BUILD = $(GO_CMD) build
GO_GET = $(GO_CMD) get

all: proxy

proxy: deps proxy.go
	$(GO_BUILD) proxy.go

deps:
	$(GO_GET) -u github.com/gorilla/mux

.PHONY: clean

clean:
	rm -f proxy
	rm -rf pkg/
	rm -rf src/
