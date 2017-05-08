PACKAGE = github.com/thetorproject/proteus
VERSION="0.1.0-dev"
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`
LDFLAGS = -ldflags "-X ${PACKAGE}/proteus-common.CommitHash=${COMMIT_HASH} -X ${PACKAGE}/proteus-common.BuildDate=${BUILD_DATE}"
NOGI_LDFLAGS = -ldflags "-X ${PACKAGE}/proteus-common.BuildDate=${BUILD_DATE}"
ARCH_LIST="linux/amd64 darwin/amd64 linux/386"
RELEASE_OSARCH = -osarch ${ARCH_LIST}
RELEASE_OUTPUT = "dist/proteus-${VERSION}.{{.OS}}-{{.Arch}}/{{.Dir}}"

vendor:
	go get github.com/kardianos/govendor
	govendor sync proteus

build-events:
	go build ${LDFLAGS} -o bin/proteus-events proteus-events/main.go
build-notify:
	go build ${LDFLAGS} -o bin/proteus-notify proteus-notify/main.go
build-registry:
	go build ${LDFLAGS} -o bin/proteus-registry proteus-registry/main.go

proteus: vendor build-events build-registry build-notify

proteus-no-gitinfo: LDFLAGS = ${NOGI_LDFLAGS}
proteus-no-gitinfo: vendor proteus

release:
	go get github.com/mitchellh/gox
	mkdir -p ./dist
	rm -rf ./dist/*
	gox ${NOGI_LDFLAGS} ${RELEASE_OSARCH} -output ${RELEASE_OUTPUT} ./proteus-events
	gox ${NOGI_LDFLAGS} ${RELEASE_OSARCH} -output ${RELEASE_OUTPUT} ./proteus-notify
	gox ${NOGI_LDFLAGS} ${RELEASE_OSARCH} -output ${RELEASE_OUTPUT} ./proteus-registry
	for x in "${ARCH_LIST}";do ARCH=$$(echo $$x | sed "s/\//-/");cp LICENSE dist/proteus-${VERSION}.$$ARCH/;tar -cvf dist/proteus-${VERSION}.$$ARCH.tar.gz -C ./dist/ proteus-${VERSION}.$$ARCH/;done

.PHONY: vendor build build-events build-notify build-registry release
