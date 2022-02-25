#! /bin/bash
go generate ./... && \
go test ./... && \
go-licenses save --save_path legal --force main.go && \
docker build --tag=ngergs/namespace-label-webhook . && \
docker push ngergs/namespace-label-webhook
