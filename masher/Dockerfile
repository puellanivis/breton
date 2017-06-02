# This is a comment
FROM golang:1.8
MAINTAINER Cassondra Foesch <puellanivis@gmail.com>
ARG BRANCH="testing"
ARG HEAD="refs/head/testing"
ARG PROJECT=
RUN apt-get update && apt-get install -y --no-install-recommends \
    bsdtar \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /bin && \
    cd /bin && \
    curl -sS -L https://github.com/google/protobuf/releases/download/v3.0.0-beta-3/protoc-3.0.0-beta-3-linux-x86_64.zip | \
        bsdtar -xvf- --exclude=readme.txt && \
    chmod 755 /bin/protoc

RUN go get -v -u \
    github.com/golang/protobuf/proto \
    github.com/golang/protobuf/protoc-gen-go \
    golang.org/x/crypto/pkcs12 \
    golang.org/x/net/idna \
    golang.org/x/oauth2/jwt \
    golang.org/x/oauth2/google \
    golang.org/x/crypto/ssh \
    golang.org/x/crypto/ssh/agent \
    golang.org/x/crypto/ssh/terminal \
    golang.org/x/net/ipv4 \
    google.golang.org/grpc \
    github.com/golang/glog \
    github.com/fsnotify/fsnotify \
    golang.org/x/net/context # intentionally last

COPY mash.sh /bin/mash.sh
WORKDIR /go/src
