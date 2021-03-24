#!/bin/bash -x
set -e

OUT_DIR=./pb

GOGOPROTO_ROOT="$(GO111MODULE=on go list -m -f '{{ .Dir }}' -m github.com/gogo/protobuf)"
GOGOPROTO_PATH="${GOGOPROTO_ROOT}:${GOGOPROTO_ROOT}/protobuf"

rm -rf $OUT_DIR/*.{go,json}

protoc --gogofast_out=plugins=grpc:. \
    -I=. \
    -I="${GOGOPROTO_PATH}" \
    pb/hello.proto
