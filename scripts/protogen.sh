set -e

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

export PATH="$PATH:$(go env GOPATH)/bin"

protoc \
  --proto_path=./proto \
  --go_out=./src/libs/proto \
  --go_opt=paths=source_relative \
  --go-grpc_out=./src/libs/proto \
  --go-grpc_opt=paths=source_relative \
  ./proto/*
