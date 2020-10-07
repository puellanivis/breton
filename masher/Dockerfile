# This docker image builds upon the golang docker image to support a more automated process for building
# golang files.
FROM golang:1.15
MAINTAINER Cassondra Foesch <puellanivis@gmail.com>

# Setup various environment and argument settings which define the building environment.
ARG BRANCH="testing"
ARG HEAD="refs/head/testing"
ARG PROJECT=
WORKDIR /go/src

# Update pkgs and pull down:
#  * bsdtar		(bsdtar handles .zip files as well as tarballs)
#  * ca-certificates	(some packages may be retrieved by https)
#  * fakeroot           (used by debian package builder)
#  * lintian            (a linter for debian package definitions)
#  * tzdata             (timezone data)
RUN apt-get update && apt-get install -y --no-install-recommends \
	bsdtar \
	ca-certificates \
	fakeroot \
	lintian \
	tzdata \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

# Here, we pull down a protoc 3 release, and unpack it.
# This _could_ technically be removed now that debian stretch (baseline since golang:1.9 has a protoc v3.0 or higher…
# However, I don’t think it makes sense to use older version, better to be able to advance as it releases outselves.
RUN mkdir -p /usr/bin && \
	cd /usr && \
	curl -sS -L https://github.com/google/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip | \
		bsdtar -xvf- --exclude=readme.txt && \
	chmod 755 /usr/bin/protoc

# Here we pull down packages which are nearly universally a good idea to have.
# * protoc	because protobuffers are a good language-neutral data-storage definition language.
# * goimports	permits us to use goimports.
# * godoc	go stopped bundling godoc with the central binary, so we have to grab it ourselves.
# * golint	permits us to use golinter.
RUN go get -v -u \
	github.com/golang/protobuf/protoc-gen-go \
	golang.org/x/tools/cmd/goimports \
	golang.org/x/tools/cmd/godoc \
	golang.org/x/lint/golint # new intentionally last

# This script is violitile, and rebuilding/retrieving every/any libraries anytime it changes is not a good idea.
COPY mash.sh /bin/mash.sh
