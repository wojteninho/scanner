#!/usr/bin/env bash

set -e

echo 'mode: atomic' > coverage.txt
GO_PACKAGES=$(go list ./...)

for PACKAGE in ${GO_PACKAGES}; do
    go test -tags=unit -race -coverpkg=./... -coverprofile=coverage.tmp -covermode=atomic ${PACKAGE}
    if [ -f coverage.tmp ]; then
        tail -n +2 coverage.tmp >> coverage.txt
        rm coverage.tmp
    fi
done
