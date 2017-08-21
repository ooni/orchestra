PACKAGE = github.com/thetorproject/proteus
VERSION="0.1.0-beta.9"
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`
LDFLAGS = -ldflags "-X ${PACKAGE}/proteus-common.CommitHash=${COMMIT_HASH} -X ${PACKAGE}/proteus-common.BuildDate=${BUILD_DATE}"
NOGI_LDFLAGS = -ldflags "-X ${PACKAGE}/proteus-common.BuildDate=${BUILD_DATE}"
ARCH_LIST=linux/amd64 darwin/amd64 linux/386
TOOL_LIST=registry events notify
RELEASE_OSARCH = -osarch "${ARCH_LIST}"
OUTPUT_SUFFIX = "${VERSION}.{{.OS}}-{{.Arch}}/{{.Dir}}"

vendor:
	go get github.com/kardianos/govendor
	govendor sync proteus

vendor-fetch:
	govendor fetch +external

fmt:
	gofmt -s -w $(GOFILES)

bindata:
	go get -u github.com/jteeuwen/go-bindata/...
	@for tool in ${TOOL_LIST}; do                                          \
	  if [ -d proteus-$$tool/data ]; then                                  \
	    extra_dirs="proteus-$$tool/data/...";                              \
	  fi;                                                                  \
	  go-bindata -prefix proteus-$$tool/                                   \
	    -o proteus-$$tool/$$tool/bindata.go -pkg $$tool                    \
	    proteus-common/data/... $$extra_dirs;                              \
	done

build-all: bindata build-events build-notify build-registry

build-events:
	go build ${LDFLAGS} -o bin/proteus-events proteus-events/main.go
build-notify:
	go build ${LDFLAGS} -o bin/proteus-notify proteus-notify/main.go
build-registry:
	go build ${LDFLAGS} -o bin/proteus-registry proteus-registry/main.go

proteus: vendor build-all

proteus-no-gitinfo: LDFLAGS = ${NOGI_LDFLAGS}
proteus-no-gitinfo: vendor proteus

release: bindata
	go get github.com/mitchellh/gox
	mkdir -p ./dist
	rm -rf ./dist/*
	gox ${NOGI_LDFLAGS} ${RELEASE_OSARCH} -output dist/proteus-events-${OUTPUT_SUFFIX} ./proteus-events
	gox ${NOGI_LDFLAGS} ${RELEASE_OSARCH} -output dist/proteus-notify-${OUTPUT_SUFFIX} ./proteus-notify
	gox ${NOGI_LDFLAGS} ${RELEASE_OSARCH} -output dist/proteus-registry-${OUTPUT_SUFFIX} ./proteus-registry
	for tool in ${TOOL_LIST};do for x in ${ARCH_LIST};do ARCH=$$(echo $$x | sed "s/\//-/");cp LICENSE dist/proteus-$$tool-${VERSION}.$$ARCH/;tar -cvf dist/proteus-$$tool-${VERSION}.$$ARCH.tar.gz -C ./dist/ proteus-$$tool-${VERSION}.$$ARCH/;done;done

.PHONY: vendor build build-events build-notify build-registry release bindata build-all fmt
