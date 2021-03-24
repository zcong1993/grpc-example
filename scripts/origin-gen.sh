#!/bin/bash -x
set -e

OUT_DIR=./pb

rm -rf $OUT_DIR/*.{go,json}

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pb/origin-hello.proto
