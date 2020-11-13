#!/usr/bin/env bash

go build
rm -rf build
mkdir build
cp Dockerfile kds build/
cd build
docker build -t kuchain/kds .
