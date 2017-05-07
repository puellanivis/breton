.PHONY: all clean install container push
TIMESTAMP=$(shell date +"%Y%m%d.%H%M")
ifneq "$(DOCKER_REGISTRY)" ""
REGISTRY=$(DOCKER_REGISTRY)/
endif
PROJECT=$(shell basename "${PWD}")
HEAD=$(shell git symbolic-ref HEAD 2> /dev/null || git rev-parse HEAD)
BRANCH=$(shell basename ${HEAD})
USER=$(shell whoami)
REPO=${REGISTRY}${USER}/${PROJECT}
BUILDNO=$(shell docker images ${REPO} | grep -c ${TIMESTAMP})

all: container

clean:
	docker rmi --force `docker images -q ${REPO}`

install:
	mkdir -p $(SNOW_MESH_ROOT)/bin
	cp masher $(SNOW_MESH_ROOT)/bin/

container:
	docker build --rm --build-arg BRANCH=${BRANCH} --build-arg PROJECT=${PROJECT} --build-arg HEAD=${HEAD} -t ${REPO}:${TIMESTAMP}-${BUILDNO} -t ${REPO}:${BRANCH} -t ${REPO}:latest .

push:
	docker push ${REPO}:${BRANCH}
