.PHONY: all test

all: build

prepare:
	go get github.com/tools/godep
	${GOPATH}/bin/godep restore

build:
#	go build -o ${GOPATH}/bin/kauthz-ars rest/authz_check/server/server.go
	go install Akso/meta Akso/store Akso/rest       
	go build -o ${GOPATH}/bin/food-rest rest/server/server.go

test: 
	go test ${TEST_OPTS} Akso/store

