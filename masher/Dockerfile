# This docker image builds upon the golang docker image to support a more automated process for building
# golang files.
FROM golang:1.9
MAINTAINER Cassondra Foesch <puellanivis@gmail.com>

# Setup various environment and argument settings which define the building environment.
ARG BRANCH="testing"
ARG HEAD="refs/head/testing"
ARG PROJECT=
WORKDIR /go/src

# Update pkgs and pull down:
#  * bsdtar		(bsdtar handles .zip files as well as tarballs)
#  * ca-certificates	(some packages may be retrieved by https)
RUN apt-get update && apt-get install -y --no-install-recommends \
	bsdtar \
	ca-certificates \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

# Here, we pull down a protoc 3 release, and unpack it.
# This _could_ technically be removed now that debian stretch (baseline since golang:1.9 has a protoc v3.0 or higher…
# However, I don’t think it makes sense to use older version, better to be able to advance as it releases outselves.
RUN mkdir -p /usr/bin && \
	cd /usr && \
	curl -sS -L https://github.com/google/protobuf/releases/download/v3.5.1/protoc-3.5.1-linux-x86_64.zip | \
		bsdtar -xvf- --exclude=readme.txt && \
	chmod 755 /usr/bin/protoc

# Here we pull down packages which are nearly universally a good idea to have.
# * x/net/context	this is so widely used, and necessary that it is intentionall placed last
#			so that the command-line will always be valid as long as previous lines end with "\\\n"
#
# * glog		because the Google logging style is nice.
# * protoc		because protobuffers are a good language-neutral data-storage definition language
# * grpc		because µservices are nice.
# * crypto/pkcs12	for supporting https
# * crypto/ssh/..	lib/files/sftp will depend upon these, plus they’re nice to have around
# * net/idna		international DNS
# * net/ipv4		provides more raw access to ipv4 packets (to better support UDT)
# * oauth2/..		OAuth2 support is nice
RUN go get -v -u \
	github.com/golang/glog \
	github.com/golang/protobuf/proto \
	github.com/golang/protobuf/protoc-gen-go \
	google.golang.org/grpc \
	golang.org/x/crypto/pkcs12 \
	golang.org/x/crypto/ssh \
	golang.org/x/crypto/ssh/agent \
	golang.org/x/crypto/ssh/terminal \
	golang.org/x/net/idna \
	golang.org/x/net/ipv4 \
	golang.org/x/oauth2/jwt \
	golang.org/x/oauth2/google \
	golang.org/x/net/context # intentionally last

# These are useful libraries that are not as centrally/clearly useful as the various Google-supported libraries.
# * fsnotify		to allow for filesystem notifications
# * yaml		YAML is really common, and this implementation is well designed.
# * gocql		Cassondra database access
# * prometheus		Prometheus metrics library (currently wrapped to a better API by breton’s lib/metrics)
RUN go get -v -u \
	github.com/fsnotify/fsnotify \
	github.com/pkg/errors \
	gopkg.in/yaml.v2 \
	github.com/gocql/gocql \
	github.com/prometheus/client_golang/prometheus \
	github.com/prometheus/client_golang/prometheus/promhttp

# These are additional libraries that can basically be localized. This should be pulled in by a different
# Dockerfile that uses FROM this image. But I’ve yet to build out this feature.
RUN go get -v -u \
	gopkg.in/natefinch/lumberjack.v2 \
	github.com/stackimpact/stackimpact-go \
	gopkg.in/tylerb/graceful.v1

# This script is violitile, and rebuilding/retrieving every/any libraries anytime it changes is not a good idea.
COPY mash.sh /bin/mash.sh
