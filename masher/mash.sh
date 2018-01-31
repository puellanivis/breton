#!/bin/bash

while [[ "$#" -gt 0 ]]; do
	key="$1"
	val="${key#*=}"

	case $key in
		--cache=*)
			export GOCACHE="$val"
		;;
		--timestamp=*)
			TIMESTAMP="$val"
		;;
		--id=*)
			ID="$val"
		;;

		--linux)
			LINUX="true"
		;;
		--darwin)
			DARWIN="true"
		;;
		--windows)
			WINDOWS="true"
		;;

		--allprotos)
			ALLPROTOS="true"
		;;
		--proto)
			NOTESTING="true"
			NOBUILD="true"
		;;

		--test)
			TESTING="true"
		;;
		--notest)
			NOTESTING="true"
		;;
		--nobuild)
			NOBUILD="true"
		;;

		--)
			shift
			break
		;;
		--shell)
			ESCAPE="true"
		;;
		--*)
			echo "$0:" "unknown flag $1" 1>&2
			exit 1
		;;
		*)
			break
		;;
	esac
	shift
done

if [[ $ESCAPE == "true" ]]; then
	exec /bin/bash "$@"
	exit 1
fi

if which go > /dev/null 2>&1 ; then
        GO_VERSION="$(go version)"
        GO_VERSION=${GO_VERSION#go version go}
        GO_VERSION=${GO_VERSION%% *}
        echo Building with Go Version: $GO_VERSION
fi

if ! ls *.go > /dev/null 2>&1 ; then
	NOGOFILES=true
fi

PROTOC_FLAGS="--proto_path=./proto --proto_path=/go/src --go_out=plugins=grpc:proto/"

if [[ -n $ALLPROTOS ]]; then
	if [[ -n $NOGOFILES ]]; then
		echo Building all subdir protos
		protos=$(find . -type d -name proto)
		proto_prefix=""
	else
		echo Building go dependency protos
		protos=$(go list -f "{{.Deps}}" | grep -o -e " [^\.\/ ]*/[^ ]*/proto\> | sort -u")
		proto_prefix=/go/src
	fi

	for proto in $protos; do
		protopath=${proto_prefix}${proto%/*}
		for protofile in $( find $protopath/proto -maxdepth 1 -name "*.proto" ); do
			echo Building all protos: $protofile
			( cd $protopath ; protoc ${PROTOC_FLAGS} ${protofile#$protopath/} ) || exit 1
		done
	done
elif [[ -d "proto" ]]; then
	echo Building local protos
	protopath=${PWD}
	for protofile in $( find $protopath/proto -maxdepth 1 -name "*.proto" ); do
		echo Building proto: $protofile
		( cd $protopath ; protoc ${PROTOC_FLAGS} ${protofile#$protopath/} ) || exit 1
	done
fi


if [[ -n $NOGOFILES ]]; then
	if [[ -x test.sh ]] ; then
		# if there is a test.sh file, then execute this instead of
		# erroring out that there are no go files.
		exec ./test.sh
	fi

	echo No go files found, not building
	exit 0
fi

case "${LINUX}${DARWIN}${WINDOWS}" in
	"")
		LINUX="true"
	;;
	truetrue*)
		TESTING="true"
esac

if [[ ("$TESTING" == "true") && ("$NOTESTING" != "true") ]]; then
	echo testing...
	go test || exit 1
fi

PACKAGE=`go list -f {{.Name}}`
if [[ $PACKAGE != "main" ]]; then
	echo This is not a package main go project.

	if [[ $PROD == "true" ]]; then
		echo Building is being aborted.
		exit 1
	fi

	echo Building is being skipped.
	exit 0
fi

echo getting dependencies...
go get -v -d || exit 1

if [[ "$NOBUILD" == "true" ]]; then
	exit 0
fi

BUILDSTAMP="$TIMESTAMP"
if [[ -n $ID ]]; then
        [[ -n $BUILDSTAMP ]] && BUILDSTAMP="${BUILDSTAMP}~"
	BUILDSTAMP="${BUILDSTAMP}${ID}"
fi

PROJECT="${PWD##*/}"
if [[ -n $BUILDSTAMP ]]; then
	DEPS=$( go list -f "{{.Deps}}" | grep -c -e "\<lib/util\>" )
	if [[ $DEPS -ne 0 ]]; then
		GOFLAGS=-ldflags="-X github.com/puellanivis/breton/lib/util.BUILD=$BUILDSTAMP"
	else
		GOFLAGS=-ldflags="-X main.VersionBuild=$BUILDSTAMP"
	fi
fi

touch /tmp/stuff > /dev/null 2>&1 || mount -t tmpfs -o rw,nodev,nosuid,size=1G /dev/null /tmp

if [[ $LINUX == "true" ]]; then
	OUT="bin/linux.x86_64"
	echo Compiling ${OUT}/${PROJECT}
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=linux GOARCH=amd64 go build -o "${OUT}/${PROJECT}" "${GOFLAGS}" || exit 1
fi

if [[ $DARWIN == "true" ]]; then
	OUT="bin/darwin.x86_64"
	echo Compiling ${OUT}/${PROJECT}
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=darwin GOARCH=amd64 go build -o "${OUT}/${PROJECT}" "${GOFLAGS}" || exit 1
fi

if [[ $WINDOWS == "true" ]]; then
	OUT="bin/windows.x86_64"
	echo Compiling ${OUT}/${PROJECT}.exe
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=windows GOARCH=amd64 go build -o "${OUT}/${PROJECT}.exe" "${GOFLAGS}" || exit 1
fi

echo Complete
