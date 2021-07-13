#!/bin/bash
# 在当前目录生成route.pb.go以及在当前目录生成route_grpc.pb.go
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative route.proto