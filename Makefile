PACKAGE = github.com/thetorproject/proteus
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`
LDFLAGS = -ldflags "-X ${PACKAGE}/proteus-common.CommitHash=${COMMIT_HASH} -X ${PACKAGE}/proteus-common.BuildDate=${BUILD_DATE}"

.PHONY: vendor build build-events build-notify build-registry

vendor:
	go get github.com/kardianos/govendor
	govendor sync proteus

build-events:
	go build ${LDFLAGS} -o bin/proteus-events proteus-events/main.go
build-notify:
	go build ${LDFLAGS} -o bin/proteus-notify proteus-notify/main.go
build-registry:
	go build ${LDFLAGS} -o bin/proteus-registry proteus-registry/main.go

build: build-events build-registry build-notify
