FROM --platform=$BUILDPLATFORM golang:1.16-alpine as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apk add --no-cache git

COPY ./ /go/src/github.com/meyskens/sniproxy

WORKDIR /go/src/github.com/meyskens/sniproxy

RUN export GOARM=6 && \
    export GOARCH=amd64 && \
    if [ "$TARGETPLATFORM" == "linux/arm64" ]; then export GOARCH=arm64; fi && \
    if [ "$TARGETPLATFORM" == "linux/arm" ]; then export GOARCH=arm; fi && \
    go build -ldflags "-X main.version=$(git rev-parse --short HEAD)" ./cmd/sniproxy/

FROM alpine:3.13

RUN apk add --no-cache ca-certificates

COPY --from=build /go/src/github.com/meyskens/sniproxy/sniproxy /usr/local/bin/

RUN mkdir /opt/sniproxy
WORKDIR /opt/sniproxy

ENTRYPOINT [ "/usr/local/bin/sniproxy" ]