#!/bin/bash -x
set -e

go install github.com/gogo/protobuf/protoc-gen-gogofast@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
