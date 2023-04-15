FROM golang:1.20-alpine as build-container
COPY . /root/app
WORKDIR /root

RUN apk --no-cache add git && \
  go install github.com/google/go-licenses@latest && \
  cd app && \
  CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -a --ldflags '-s -w' && \
  go-licenses save ./... --save_path=legal

FROM gcr.io/distroless/static
COPY --from=build-container /root/app/k8s-adm-ctrl /app/k8s-adm-ctrl
COPY --from=build-container /root/app/legal /app/legal
USER 1000
EXPOSE 10250
ENTRYPOINT ["/app/k8s-adm-ctrl","-port","10250"]
CMD []
