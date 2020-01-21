PACKAGE = github.com/ooni/orchestra
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`
LDFLAGS = -ldflags "-X ${PACKAGE}/common.CommitHash=${COMMIT_HASH} -X ${PACKAGE}/common.BuildDate=${BUILD_DATE}"
NOGI_LDFLAGS = -ldflags "-X ${PACKAGE}/common.BuildDate=${BUILD_DATE}"
TOOL_LIST = registry orchestrate
GOPATH ?= `go env GOPATH`

vendor: vendor-fetch
vendor-fetch:

fmt:
	go fmt ./...
fmt-check:
	test -z "$$(go fmt ./...)"

lint:
	go get -v golang.org/x/lint/golint@latest
	${GOPATH}/bin/golint -set_exit_status ./...

test: fmt-check
	go get -v golang.org/x/tools/cmd/cover@latest
	go get -v github.com/mattn/goveralls@latest
	go test -coverprofile=coverage.txt -coverpkg=./... ./...
	@echo "Consider: go tool cover -html=coverage.txt"

check: fmt-check test

bindata:
	go get -v github.com/shuLhan/go-bindata/cmd/go-bindata@v3.4.0
	@${GOPATH}/bin/go-bindata                                                   \
		-nometadata                                                         \
		-o common/bindata.go -pkg common                                    \
			common/data/...;
	@for tool in ${TOOL_LIST}; do                                               \
		if [ -d $$tool/data ]; then                                         \
			extra_dirs="$$tool/data/...";                               \
		fi;                                                                 \
		${GOPATH}/bin/go-bindata                                            \
			-nometadata                                                 \
			-o $$tool/$$tool/bindata.go -pkg $$tool                     \
			common/data/... $$extra_dirs;                               \
	done
	@for bindatafile in $$(find . -type f -name bindata.go); do                 \
		go fmt $$bindatafile;                                               \
	done

build-all: bindata build-orchestrate build-registry

build-orchestrate:
	go build ${LDFLAGS} -o bin/ooni-orchestrate orchestrate/main.go
build-registry:
	go build ${LDFLAGS} -o bin/ooni-registry registry/main.go

orchestra: vendor build-all

orchestra-no-gitinfo: LDFLAGS = ${NOGI_LDFLAGS}
orchestra-no-gitinfo: vendor orchestra

dirty-check:
	# Until https://github.com/golang/go/issues/30515 lands, we need to
	# prevent go get from changing go.mod, or goreleaser punts. So we
	# make sure there are no changes in tree manually first. Then we make
	# sure we have goreleaser. Finally we reset go.mod go.sum changes.
	git diff --exit-code --quiet

release: dirty-check fmt-check bindata
	go get -u -v github.com/goreleaser/goreleaser
	git checkout go.mod go.sum  # See dirty-check's comment
	# XXX Not a fan of how it autogens the release notes, we should probably
	# extract them from our changelog and embed them using:
	# https://goreleaser.com/#releasing.custom_release_notes
	GITHUB_TOKEN=`cat .GITHUB_TOKEN` ${GOPATH}/bin/goreleaser --rm-dist

.PHONY: vendor build build-orchestrate build-registry release bindata build-all fmt fmt-check check dirty-check
