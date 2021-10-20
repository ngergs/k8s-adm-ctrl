#! /bin/bash
go generate ./... && \
go test ./... && \
go-licenses save --save_path legal --force main.go && \
docker build --tag=selfenergy/namespace-label-webhook . && \
docker push selfenergy/namespace-label-webhook
