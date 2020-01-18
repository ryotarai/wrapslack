export GO111MODULE=on
export VERSION=$(shell git describe --tags | grep ^v)

.PHONY: build
build:
	go build -o bin/wrapslack .

release:
	./scripts/release.sh
