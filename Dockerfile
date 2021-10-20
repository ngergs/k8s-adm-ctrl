FROM golang:alpine
COPY . /root/
WORKDIR /root
RUN CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -a --ldflags '-s -w'

FROM alpine:latest
RUN apk --no-cache add ca-certificates && mkdir /root/legal
WORKDIR /root/
COPY --from=0 /root/k8s-admission-ctrl .
COPY legal legal
EXPOSE 8080
ENTRYPOINT ["./k8s-admission-ctrl"]
CMD []
