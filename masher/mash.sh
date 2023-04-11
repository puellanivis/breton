#!/bin/bash

ARCH="amd64"
ARCH_DIR="x86_64"

while [[ $# -gt 0 ]]; do
	key="$1"
	val="${key#*=}"

	case "$key" in
	--cache=*)
		export GOCACHE="$val"
	;;
	--buildver=*)
		BUILDVER="$val"
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

	--arm64)
		ARCH="arm64"
		ARCH_DIR="arm64"
	;;
	--linux)
		LINUX="true"
	;;
	--darwin)
		DARWIN="true"
	;;
	--openbsd)
		OPENBSD="true"
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

	--shell)
		ESCAPE="true"
	;;

	--)
		shift
		break
	;;
	--*)
		echo "$0: unknown flag $key" 1>&2
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

case "${LINUX}${DARWIN}${OPENBSD}${WINDOWS}" in
"")
	if [[ -n $DEB ]]; then
		LINUX="true"
	else
		NOCOMPILE="true"
	fi
;;
truetrue*)
	TESTING="true"
esac

if which go > /dev/null 2>&1 ; then
        GO_VERSION="$(go version)"
        GO_VERSION=${GO_VERSION#go version go}
        GO_VERSION=${GO_VERSION%% *}
        [[ -z $NOCOMPILE ]] && echo "Building with Go Version: $GO_VERSION"
fi

if ! ls *.go > /dev/null 2>&1 ; then
	NOGOFILES=true
fi

PROTOC_FLAGS="--go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative"

if [[ -n $ALLPROTOS ]]; then
	if [[ -n $NOGOFILES ]]; then
		echo "Building all subdir protos"
		protos=$( find . -type f -name "*.proto" -exec dirname \{\} \; | sort -u )
	else
		echo "Building go dependency protos"
		protos=$( go list -f "{{range .Deps}}{{println}}{{end}}" | grep "/proto$" | sort -u )
	fi

	for proto in $protos; do
		echo "Building proto: $proto"
		protoc ${PROTOC_FLAGS} "${proto#./}/"*.proto || exit 1
	done

elif [[ -d "proto" ]]; then
	echo "Building local proto"
	protoc ${PROTOC_FLASG} proto/*.proto
fi


if [[ -n $NOGOFILES ]]; then
	if [[ -x test.sh ]] ; then
		# if there is a test.sh file, then execute this instead of
		# erroring out that there are no go files.
		exec ./test.sh
	fi

	echo "No go files found, not building"
	exit 0
fi

if [[ ("$TESTING" == "true") && ("$NOTESTING" != "true") ]]; then
	echo "Testing..."
	go test || exit 1
fi

PACKAGE=`go list -f {{.Name}}`
if [[ $PACKAGE != "main" ]]; then
	echo "This is not a package main go project."

	if [[ $PROD == "true" ]]; then
		echo "Building is being aborted."
		exit 1
	fi

	echo "Building is being skipped."
	exit 0
fi

if [[ $VENDOR != "true" && $NOCOMPILE != "true" ]]; then
	echo "Getting dependencies..."
	if [[ -r go.mod ]]; then
		echo "Using go modules..."

	elif [[ -r Gopkg.toml ]]; then
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
	LDFLAGS="-ldflags=-X main.VersionBuild=$BUILDSTAMP -X main.Buildstamp=$BUILDSTAMP"
	if [[ -n $BUILDVER ]]; then
		LDFLAGS="$LDFLAGS -X main.Version=$BUILDVER"
	fi

	LIB_UTIL_DEP=$( go list -f "{{range .Deps}}{{println}}{{end}}" | grep -e "\<breton/lib/util\>" | head -n1 )
	if [[ -n $LIB_UTIL_DEP ]]; then
		LDFLAGS="$LDFLAGS -X ${LIB_UTIL_DEP}.BUILD=$BUILDSTAMP"
	fi
fi

touch /tmp/stuff > /dev/null 2>&1 || mount -t tmpfs -o rw,nodev,nosuid,size=1G /dev/null /tmp

if [[ -r ./.gitignore ]]; then
	grep -qFx -e "bin" ./.gitignore || echo "bin" >> ./.gitignore
fi

if [[ $LINUX == "true" ]]; then
	OUT="bin/linux.${ARCH_DIR}"
	echo "Compiling ${OUT}/${PROJECT}"
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=linux GOARCH=$ARCH go build -o "${OUT}/${PROJECT}" "${LDFLAGS}" || exit 1

	[[ ( ${DEB} != "false" ) && ( -d debian )  ]] && DEB="true"
fi

if [[ $DARWIN == "true" ]]; then
	OUT="bin/darwin.${ARCH_DIR}"
	echo "Compiling ${OUT}/${PROJECT}"
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=darwin GOARCH=$ARCH go build -o "${OUT}/${PROJECT}" "${LDFLAGS}" || exit 1
fi

if [[ $OPENBSD == "true" ]]; then
	OUT="bin/openbsd.${ARCH_DIR}"
	echo "Compiling ${OUT}/${PROJECT}"
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=openbsd GOARCH=$ARCH go build -o "${OUT}/${PROJECT}" "${LDFLAGS}" || exit 1
fi

if [[ $WINDOWS == "true" ]]; then
	OUT="bin/windows.${ARCH_DIR}"
	echo "Compiling ${OUT}/${PROJECT}.exe"
	[ -d "$OUT" ] || mkdir -p $OUT || exit 1
	GOOS=windows GOARCH=$ARCH go build -o "${OUT}/${PROJECT}.exe" "${LDFLAGS}" || exit 1
fi

if [[ ( $DEB == "true" ) && ( -x bin/linux.${ARCH_DIR}/${PROJECT} ) ]]; then
	[ -d debian ] || mkdir debian

	echo "Building debian package..."

	VERSION="${BUILDSTAMP}"
	BIN_VERSION="$( ./bin/linux.${ARCH_DIR}/${PROJECT} --version )"
	[[ -n $BIN_VERSION ]] && echo "BIN_VERSION=\"$BIN_VERSION\""
	case "${BIN_VERSION}" in
	*\ v*)
		VERSION="$( echo "${BIN_VERSION}" | cut -d" " -f2 )"
		VERSION="${VERSION#v}"
	;;
	esac
	echo "VERSION=${VERSION}"

	if [[ -r ./.gitignore ]]; then
		grep -qFx -e "build" ./.gitignore || echo "build" >> ./.gitignore
		grep -qFx -e "debian/changelog.gz" -e "*.gz" ./.gitignore || echo "debian/changelog.gz" >> ./.gitignore
		grep -qFx -e "*.deb" ./.gitignore || echo "*.deb" >> ./.gitignore
	fi

	WHOAMI="nobody <nobody@example.com>"
	[[ -r debian/whoami ]] && WHOAMI="$(cat debian/whoami)"

	ORIGIN_URL="$( git remote get-url origin )"


	if [[ ! -r debian/control ]]; then
		cat <<EOF > debian/control
Source: @@PROJECT@@
Section: unknown
Priority: optional
Maintainer: $WHOAMI
Homepage: $ORIGIN_URL
Package: @@PROJECT@@
Architecture: @@ARCH@@
Version: @@VERSION@@
Description: TODO
EOF
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

	DEB_PACKAGE="$( grep "^Package: " build/DEBIAN/control | cut -d" " -f2 )"

	# install copyright and changelog documentation
	install -d "build/usr/share/doc/${DEB_PACKAGE}"

	if [[ ! -r debian/copyright ]]; then
		if [[ -r LICENSE ]]; then
			LICENSE="LICENSE"
		elif [[ -r LICENSE.md ]]; then
			LICENSE="LICENSE.md"
		fi

		YEARS="$( date +%Y )"
		ORIGIN_URL="$( git remote get-url origin )"

		if [[ -n $LICENSE ]]; then (
			cat <<EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: $PROJECT
Upstream-Contact: $WHOAMI
Source: $ORIGIN_URL

Files: *
Copyright: $YEAR $WHOAMI
License: LICENSE

License: LICENSE
EOF
			awk '!NF{$0="."}1' "$LICENSE"
		) > debian/copyright; fi
	fi

	if [[ -r debian/copyright ]]; then
		install -m644 debian/copyright "build/usr/share/doc/${DEB_PACKAGE}/"
	fi

	if [[ -r CHANGELOG ]]; then
		# if CHANGELOG is newer than changelog.gz, then compress it and write it to changelog.gz
		if [[ CHANGELOG -nt debian/changelog.gz ]]; then
			gzip -9 -c CHANGELOG > debian/changelog.gz
		fi
	else

		# if changelog.gz does not exist, then make it with a bare-minimum changelog gzip file.
		cat <<EOF | gzip -9 > debian/changelog.gz
${DEB_PACKAGE} (${VERSION}) unstable; urgency=low

  * No information.

 -- ${WHOAMI}  $(date -R)
EOF
	fi

	install -m644 debian/changelog.gz "build/usr/share/doc/${DEB_PACKAGE}/changelog.Debian.gz"

	# install binary
	install -d build/usr/bin
	install -m755 "bin/linux.${ARCH_DIR}/${PROJECT}" "build/usr/bin/${PROJECT}"

	if [[ ( -n ${DEB_PACKAGE} ) && ( -n ${VERSION} ) && ( -n ${ARCH} ) ]]; then
		DEB_FILE="${DEB_PACKAGE}_${VERSION}_${ARCH}.deb"

		# Final packaging
		rm -f "${DEB_FILE}"
		fakeroot dpkg-deb --build build .
		lintian -X binaries "${DEB_FILE}"
	fi
fi

echo "Complete"
