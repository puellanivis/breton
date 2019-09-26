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
		--private=*)
			export GOPRIVATE="$val"
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

		--deb)
			DEB="true"
		;;
		--nodeb)
			DEB="false"
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

if [[ -d ./vendor ]]; then
	export GOFLAGS="-mod=vendor"
	VENDOR="true"
fi

if [[ $ESCAPE == "true" ]]; then
	exec /bin/bash "$@"
	exit 1
fi

case "${LINUX}${DARWIN}${WINDOWS}" in
	"")
		[[ $DEB == "" ]] && LINUX="true"
	;;
	truetrue*)
		TESTING="true"
esac

[[ "${LINUX}${DARWIN}${WINDOWS}" == "" ]] && NOCOMPILE="true"

if which go > /dev/null 2>&1 ; then
        GO_VERSION="$(go version)"
        GO_VERSION=${GO_VERSION#go version go}
        GO_VERSION=${GO_VERSION%% *}
        [[ -z "$NOCOMPILE" ]] && echo Building with Go Version: $GO_VERSION
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

if [[ $VENDOR != "true" && $NOCOMPILE != "true" ]]; then
	echo getting dependencies...
	if [[ -r Gopkg.toml ]]; then
		DEP_UP=""
		if [[ -r Gopkg.lock ]]; then
			DEP_UP="-update"
		fi

		dep ensure $DEP_UP
	else
		go get -v -d || exit 1
	fi
fi

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
	LDFLAGS="-ldflags=-X main.VersionBuild=$BUILDSTAMP -X main.Buildstamp=$BUILDSTAMP"
	if [[ $DEPS -ne 0 ]]; then
		LDFLAGS="$LDFLAGS -X github.com/puellanivis/breton/lib/util.BUILD=$BUILDSTAMP"
	fi
fi

touch /tmp/stuff > /dev/null 2>&1 || mount -t tmpfs -o rw,nodev,nosuid,size=1G /dev/null /tmp

if [[ -r ./.gitignore ]]; then
	grep -qFx -e "bin" ./.gitignore || echo "bin" >> ./.gitignore
fi

if [[ $LINUX == "true" ]]; then
	OUT="bin/linux.x86_64"
	echo Compiling ${OUT}/${PROJECT}
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=linux GOARCH=amd64 go build -o "${OUT}/${PROJECT}" "${LDFLAGS}" || exit 1

	[[ "$DEB" != "false" ]] && DEB="true"
fi

if [[ $DARWIN == "true" ]]; then
	OUT="bin/darwin.x86_64"
	echo Compiling ${OUT}/${PROJECT}
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=darwin GOARCH=amd64 go build -o "${OUT}/${PROJECT}" "${LDFLAGS}" || exit 1
fi

if [[ $WINDOWS == "true" ]]; then
	OUT="bin/windows.x86_64"
	echo Compiling ${OUT}/${PROJECT}.exe
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=windows GOARCH=amd64 go build -o "${OUT}/${PROJECT}.exe" "${LDFLAGS}" || exit 1
fi

if [[ ( $DEB == "true" ) && ( -r debian/control ) && ( -x bin/linux.x86_64/${PROJECT} ) ]]; then
	VERSION=${BUILDSTAMP}
	BIN_VERSION="$(./bin/linux.x86_64/${PROJECT} --version)"
	[[ -n $BIN_VERSION ]] && echo BIN_VERSION=\"$BIN_VERSION\"
	case "${BIN_VERSION}" in
	*\ v*)
		VERSION=$(echo "${BIN_VERSION}" | cut -d" " -f2)
		VERSION=${VERSION#v}
	;;
	esac
	echo VERSION=${VERSION}
	ARCH="amd64" # TODO(puellanivis): this shouldn’t be baked in, but it’s already baked in all over, already.

	if [[ -r ./.gitignore ]]; then
		grep -qFx -e "build" ./.gitignore || echo "build" >> ./.gitignore
		grep -qFx -e "debian/changelog.gz" -e "*.gz" ./.gitignore || echo "debian/changelog.gz" >> ./.gitignore
		grep -qFx -e "*.deb" ./.gitignore || echo "*.deb" >> ./.gitignore
	fi

	install -d build/DEBIAN
	for f in conffiles control; do
		if [[ -r "debian/$f" ]]; then
			install -m644 debian/$f build/DEBIAN/
			sed -i "s/@@PROJECT@@/${PROJECT}/" build/DEBIAN/$f
			sed -i "s/@@VERSION@@/${VERSION}/" build/DEBIAN/$f
			sed -i "s/@@ARCH@@/${ARCH}/" build/DEBIAN/$f
		fi
	done
	for f in postinst postrm prerm; do
		if [[ -r debian/$f ]]; then
			install -m755 debian/$f build/DEBIAN/
		fi
	done

	DEB_PACKAGE=$(grep "^Package: " build/DEBIAN/control | cut -d" " -f2 )

	# install copyright and changelog documentation
	install -d build/usr/share/doc/${DEB_PACKAGE}
	install -m644 debian/copyright build/usr/share/doc/${DEB_PACKAGE}/

	if [[ -r CHANGELOG ]]; then
		# if CHANGELOG is newer than changelog.gz, then compress it and write it to changelog.gz
		if [[ CHANGELOG -nt debian/changelog.gz ]]; then
			gzip -9 -c CHANGELOG > debian/changelog.gz
		fi
	else
		WHOAMI="nobody <nobody@example.com>"
		[[ -r debian/whoami ]] && WHOAMI="$(cat debian/whoami)"

		# if changelog.gz does not exist, then make it with a bare-minimum changelog gzip file.
		printf "${DEB_PACKAGE} (${VERSION}) unstable; urgency=low\n\n  * No information.\n\n -- ${WHOAMI}  $(date -R)\n" | gzip -9 > debian/changelog.gz
	fi
	install -m644 debian/changelog.gz build/usr/share/doc/${DEB_PACKAGE}/changelog.Debian.gz

	# install binary
	install -d build/usr/bin
	install -m755 bin/linux.x86_64/${PROJECT} build/usr/bin/${PROJECT}

	if [[ ( -n ${DEB_PACKAGE} ) && ( -n ${VERSION} ) && ( -n ${ARCH} ) ]]; then
		DEB_FILE="${DEB_PACKAGE}_${VERSION}_${ARCH}.deb"

		# Final packaging
		rm -f "${DEB_FILE}"
		fakeroot dpkg-deb --build build .
		lintian -X binaries "${DEB_FILE}"
	fi
fi

echo Complete
