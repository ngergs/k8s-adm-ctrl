FROM golang:alpine as build-container
COPY . /root/
WORKDIR /root
RUN CGO_ENABLED=0 GOOD=linux GOARCH=amd64 go build -a --ldflags '-s -w'

FROM gcr.io/distroless/static
COPY --from=build-container /root/k8s-admission-ctrl /app/k8s-admission-ctrl
COPY legal app/legal
USER 1000
EXPOSE 8080
ENTRYPOINT ["/app/k8s-admission-ctrl"]
CMD []
