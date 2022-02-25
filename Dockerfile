FROM golang:alpine as build-container
COPY . /root/
WORKDIR /root
RUN CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -a --ldflags '-s -w' && \
go get github.com/google/go-licenses && \
go-licenses save ./... --save_path=legal

FROM gcr.io/distroless/static
COPY --from=build-container /root/k8s-adm-ctrl /app/k8s-adm-ctrl
COPY --from=build-container /root/legal /app/legal
USER 1000
EXPOSE 10250
ENTRYPOINT ["/app/k8s-adm-ctrl","-port","10250"]
CMD []
