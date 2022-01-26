#!/bin/bash

export GO111MODULE=on
export CGO_ENABLED=1
export GOOS=darwin

mkdir -p ./local
rm -f ./local/SHA256
go mod download
go build -a -v -o ./local/vault-ethereum .
openssl sha256 "./local/vault-ethereum" | awk '{print $2}' > ./local/SHA256;
