#!/bin/bash -e

ROOTDIR=$(cd $(dirname $0);pwd)

proto_dir=$ROOTDIR/../examples/protobuf

cd $proto_dir

for file in $(find . -type f -name "*.proto")
do
    protoc --ts_out=${GOPATH}/src \
        --go_out=${GOPATH}/src \
        --go-grpc_out=${GOPATH}/src \
        --go-rest_out=${GOPATH}/src \
        -I${ROOTDIR}/../third_party -I. $file
done

for file in $(find . -type f -name "*.pb.go")
do
    sed -i 's/,omitempty//g' $file
done
