#!/usr/bin/env bash

set -e

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

PATH="$PATH:$(go env GOPATH)/bin"

mkdir -p ./proto/go/pb

protoc \
  --proto_path=./proto/spec \
  --go_out=./proto/go/pb \
  --go_opt=paths=source_relative \
  --go-grpc_out=./proto/go/pb \
  --go-grpc_opt=paths=source_relative \
  ./proto/spec/*
