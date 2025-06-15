#!/usr/bin/env bash
set -e
# Install the protoc-gen-go-errors-code binary in $GOPATH/bin
go build -o $GOPATH/bin/protoc-gen-go-errors-code