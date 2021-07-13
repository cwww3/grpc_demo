#!/bin/bash
# 在当前目录运行 生成route.pb.go以及route_grpc.pb.go
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative route.proto