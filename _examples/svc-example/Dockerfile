FROM golang:1.24.9 AS builder

ARG services
ARG GOBUILD_CMD
ARG GOOS
ARG CGO_ENABLED

WORKDIR /app

COPY . .


RUN make all_env &&\
    make gen_proto && \
    go mod tidy && \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=amd64 ${GOBUILD_CMD} -ldflags "-s -w" -o bin/server ./${services}

FROM scratch:latest

ARG services

WORKDIR /app

COPY --from=builder /app/bin/server /app/
COPY ./${services}/conf/ ./conf/
COPY ./conf/ ./conf/conf/

CMD [ "./server" ]
