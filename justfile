#!/usr/bin/env just --justfile

XDIR := justfile_directory()
XDIST := XDIR + "/dist"
XTMP := XDIR + "/tmp"
EXE := "gverctl"
OSARCH := os()+"-"+arch()

# GOROOT := "/u/fredo/bin/go/go1.26.7"
GOROOT := "/u/fredo/go/go1.27.4"

update-version-info:
    #!/bin/sh
    V=$(shtool version -n "{{EXE}}" -d long -l txt ./version.txt)
    echo "$V" >{{XDIR}}/version_info.txt
    echo "package pkg" >{{XDIR}}/pkg/version.go
    echo "const (" >>{{XDIR}}/pkg/version.go
    echo "    PkgVersion = \"$V\"" >>{{XDIR}}/pkg/version.go
    echo ")" >>{{XDIR}}/pkg/version.go

set-drel: do-inc-level
    #!/bin/bash
    V=$(date '+%Y.%m.')
    V=$V$(cd {{XDIR}} && shtool version -n "{{EXE}}" -l short ./version.txt|cut -f 3 -d.)
    cd {{XDIR}} && just -f justfile set-version "$V"

git-push-level: inc-level git-push

git-push: (git-push-msg "upd")
git-push-msg _MSG:
    #!/bin/sh
    VERSION=$(shtool version -l txt ./version.txt)
    TIME=$(date '+%F %T')
    git commit --all -m "{{_MSG}} - v$VERSION on $TIME"
    git push

inc-version: do-inc-version update-version-info
do-inc-version:
    #!/bin/bash
    cd {{XDIR}}
    shtool version -n "{{EXE}}" -i v -l txt ./version.txt

inc-major: do-inc-major update-version-info
do-inc-major:
    #!/bin/bash
    cd {{XDIR}}
    shtool version -n "{{EXE}}" -i r -l txt ./version.txt

inc-level: do-inc-level update-version-info
do-inc-level:
    #!/bin/bash
    cd {{XDIR}}
    shtool version -n "{{EXE}}" -i l -l txt ./version.txt

set-version _VERSION: (do-set-version _VERSION) update-version-info

do-set-version _VERSION:
    #!/bin/bash
    cd {{XDIR}}
    shtool version -n "{{EXE}}" -s "{{_VERSION}}" -l txt ./version.txt

make-update: git-push-level

make-prel: (make-prel-msg "upd")
make-prel-msg _MSG: set-drel git-push
    #!/bin/bash
    VERSION=$(shtool version -l txt ./version.txt)
    VERL=$(shtool version -l text -d long ./version.txt)
    MESSAGE="{{_MSG}} - {{EXE}} automated pre-release version $VERL"
    gh release create v$VERSION --notes "$MESSAGE" --prerelease

make-rel: (make-rel-msg "upd")
make-rel-msg _MSG: set-drel git-push
    #!/bin/bash
    VERSION=$(shtool version -l txt ./version.txt)
    VERL=$(shtool version -l text -d long ./version.txt)
    MESSAGE="{{_MSG}} - {{EXE}} automated release version $VERL"
    gh release create v$VERSION --notes "$MESSAGE"

make-upload:
    #!/bin/bash
    VERSION=$(shtool version -l txt ./version.txt)
    gh release upload v$VERSION {{XDIST}}/{{EXE}}-*

make-prel-full _MSG: (make-prel-msg _MSG) build-all make-upload
make-rel-full _MSG: (make-rel-msg _MSG) build-all make-upload

build-all:
    #!/bin/sh
    rm -rf {{XDIST}}; mkdir -p {{XDIST}}
    cd {{XDIR}} && just -f justfile build-osarch linux amd64
    cd {{XDIR}} && just -f justfile build-osarch linux arm64
    cd {{XDIR}} && just -f justfile build-osarch windows amd64
    cd {{XDIR}} && just -f justfile build-osarch windows arm64
    cd {{XDIR}} && just -f justfile build-osarch darwin amd64
    cd {{XDIR}} && just -f justfile build-osarch darwin arm64

build-all-inc:  do-inc-level build-all

build-osarch _GOS _GARCH:
    #!/bin/sh
    VERSION=$(shtool version -l txt ./version.txt)
    _gos={{_GOS}}
    _garch={{_GARCH}}
    _gexe=
    case $_gos in
        windows)
            _gexe=.exe
            ;;
        *)
            _gexe=
            ;;
    esac
    _gout={{EXE}}-${VERSION}-${_gos}-${_garch}${_gexe}
    #_gout={{EXE}}-${_gos}-${_garch}${_gexe}
    cd {{XDIR}} && just -f justfile build-osarch-out "${_gos}" "${_garch}" "${_gout}"

build-osarch-out _GOS _GARCH _GOUT: update-version-info
    #!/bin/sh
    export GOROOT={{GOROOT}}
    export PATH=$GOROOT/bin:$PATH
    mkdir -p {{XDIST}}
    #CGO_ENABLED=1 CC=musl-gcc \
    #  go build -ldflags="-linkmode external -extldflags '-static'"
    _gos={{_GOS}}
    _garch={{_GARCH}}
    _gout={{_GOUT}}
    echo "building ${_gos} ${_garch} ... ${_gout}"
    GOOS=${_gos} GOARCH=${_garch} go build -buildmode=pie -o "{{XDIST}}/${_gout}" -v main.go
