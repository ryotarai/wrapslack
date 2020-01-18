#!/bin/bash
set -eu
set -o pipefail

if [[ ! -z "$(git status -s)" ]]; then
  echo "git repo is dirty"
  exit 1
fi

rm -rf tmp/release
gox -os='linux darwin' -arch='amd64' -output="tmp/release/{{.Dir}}_${VERSION}_{{.OS}}_{{.Arch}}" -ldflags="-X main.version=${VERSION}"
gzip tmp/release/*
ghr "${VERSION}" "tmp/release/"
