#!/bin/bash
# 在当前目录运行
# 由于生成grpc-gateway和openapi需要导入第三方的proto文件(下载并保存在本地)
# 下载地址 https://github.com/googleapis/googleapis

# 通过buf可能直接使用远程的proto文件,不需要下载到本地
protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
--openapiv2_out=. --openapiv2_opt=logtostderr=true \
route.proto
