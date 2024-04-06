#!/bin/bash

PROTO_DIR=proto
OUTPUT_DIR=src/proto/lib/go
PACKAGE=$(head -1 src/go.mod | awk '{print $2}')

mkdir -p $OUTPUT_DIR
echo "compiling proto files..."
protoc -I$PROTO_DIR --go_opt=module=$PACKAGE --go_out=$OUTPUT_DIR --go-grpc_opt=module=$PACKAGE --go-grpc_out=$OUTPUT_DIR $PROTO_DIR/*.proto

echo done!
