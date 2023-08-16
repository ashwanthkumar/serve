APPNAME=serve
VERSION=0.0.1-dev
TESTFLAGS=-v -cover -covermode=atomic -bench=.

build:
	go build -tags netgo -ldflags "-w" -o ${APPNAME} .

build-linux:
	GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags "-w -s -X main.APP_VERSION=${VERSION}" -v -o ${APPNAME}-linux-amd64 .
	shasum -a256 ${APPNAME}-linux-amd64

build-mac:
	GOOS=darwin GOARCH=amd64 go build -tags netgo -ldflags "-w -s -X main.APP_VERSION=${VERSION}" -v -o ${APPNAME}-darwin-amd64 .
	shasum -a256 ${APPNAME}-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -tags netgo -ldflags "-w -s -X main.APP_VERSION=${VERSION}" -v -o ${APPNAME}-darwin-arm64 .
	shasum -a256 ${APPNAME}-darwin-arm64

build-all: build-mac build-linux

all: setup
	build
	install

setup:
	go mod download	
