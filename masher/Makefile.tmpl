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

all: Makefile Dockerfile container

Dockerfile: Dockerfile.tmpl
	cp Dockerfile.tmpl Dockerfile

Makefile: Makefile.tmpl
	mkmakefile
	exit 1

clean: Makefile
	docker rmi --force `docker images -q ${REPO}`

install: Makefile
	mkdir -p $(SNOW_MESH_ROOT)/bin
	cp masher $(SNOW_MESH_ROOT)/bin/

container: Makefile
	docker build --rm --build-arg BRANCH=${BRANCH} --build-arg PROJECT=${PROJECT} --build-arg HEAD=${HEAD} -t ${REPO}:${TIMESTAMP}-${BUILDNO} -t ${REPO}:${BRANCH} -t ${REPO}:latest .

push: Makefile
	docker push ${REPO}:${BRANCH}
