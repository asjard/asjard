#!/bin/bash -e

ROOTDIR=$(cd $(dirname $0);pwd)

proto_dir=$ROOTDIR/../core/protobuf

cd $proto_dir

for file in $(find . -type f -name "*.proto")
do
    protoc --go_out=${GOPATH}/src --go-grpc_out=${GOPATH}/src --go-rest_out=${GOPATH}/src -I${GOPATH}/src -I. $file
done
