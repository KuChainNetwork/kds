#!/usr/bin/env bash

# swag文档
swag init

# 构建
go build

# 构建镜像
rm -rf build
mkdir build
cp Dockerfile kds build/
cd build
docker build -t kuchain/kds .
