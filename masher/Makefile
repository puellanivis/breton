.PHONY: all clean install container push

TIMESTAMP:=$(shell date +"%Y%m%d.%H%M")

ifneq "$(DOCKER_REGISTRY)" ""
#REGISTRY:=$(DOCKER_REGISTRY)/
endif

PROJECT:=$(shell basename "${PWD}")
USER   :=$(shell id -un | tr '[:upper:]' '[:lower:]' | grep -o '[^\\]*$$')
REPO   :=${REGISTRY}${USER}/${PROJECT}
BUILDNO:=$(shell docker images ${REPO} | grep -c ${TIMESTAMP})

BUILDTAG:=
ifneq "$(BUILDNO)" "0"
BUILDTAG:=-$(BUILDNO)
endif

all: container

clean:
	docker rmi --force `docker images -q ${REPO}`

install:
	mkdir -p $(HOME)/bin
	cp masher $(HOME)/bin/

container:
	docker build --rm --network=host --build-arg PROJECT=${PROJECT} -t ${REPO}:${TIMESTAMP}${BUILDTAG} -t ${REPO}:latest .

push:
	[ -n "${BRANCH}" ] && docker push ${REPO}:${BRANCH} || true
	docker push ${REPO}:latest
