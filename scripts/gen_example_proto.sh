#!/bin/bash -e

ROOTDIR=$(cd $(dirname $0);pwd)

proto_dir=$ROOTDIR/../examples/protobuf

protoc_out=

go_out() {
    [[ "$protoc_out" =~ "--go_out=" ]] || protoc_out="$protoc_out --go_out=${GOPATH}/src"
}

ts_out() {
    [[ "$protoc_out" =~ "--ts_out=" ]] || protoc_out="$protoc_out --ts_out=${GOPATH}/src"
}

go_grpc_out(){
    go_out
    [[ "$protoc_out" =~ "--go-grpc_out=" ]] || protoc_out="$protoc_out --go-grpc_out=${GOPATH}/src"
}

go_rest_out() {
    go_grpc_out
    [[ "$protoc_out" =~ "--go-rest_out=" ]] || protoc_out="$protoc_out --go-rest_out=${GOPATH}/src"
}

go_rest_gw_out() {
    go_rest_out
    [[ "$protoc_out" =~ "--go-rest2grpc-gw_out=" ]] || protoc_out="$protoc_out --go-rest2grpc-gw_out=${GOPATH}/src"
}

if [  "$GEN_PROTO_GO" == "true" ];then
    go_out
fi

if [ "$GEN_PROTO_GO_GRPC" == "true" ];then
    go_grpc_out
fi

if [ "$GEN_PROTO_GO_REST" == "true" ];then
    go_rest_out
fi

if [ "$GEN_PROTO_GO_REST_GW" == "true" ];then
    go_rest_gw_out
fi

if [ "$GEN_PROTO_TS" == "true" ];then
    ts_out
fi

cd $proto_dir

for file in $(find . -type f -name "*.pb.go" -o -name '*.d.ts')
do
    rm -rf $file
done

for file in $(find . -type f -name "*.proto")
do
    protoc ${protoc_out} \
        -I${ROOTDIR}/../third_party \
        -I${ROOTDIR}/../third_party/github.com/google/gnostic \
        -I. \
        $file
done
