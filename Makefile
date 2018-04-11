PACKAGE = github.com/ooni/orchestra
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -name "bindata.go")
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_DATE = `date +%FT%T%z`
LDFLAGS = -ldflags "-X ${PACKAGE}/common.CommitHash=${COMMIT_HASH} -X ${PACKAGE}/common.BuildDate=${BUILD_DATE}"
NOGI_LDFLAGS = -ldflags "-X ${PACKAGE}/common.BuildDate=${BUILD_DATE}"
ARCH_LIST=linux/amd64 darwin/amd64 linux/386
TOOL_LIST=registry orchestrate
RELEASE_OSARCH = -osarch "${ARCH_LIST}"
GO_BINDATA_VERSION = $(shell go-bindata --version | cut -d ' ' -f2 | head -n 1 || echo "missing")
REQ_GO_BINDATA_VERSION = 3.2.0

vendor:
	dep ensure
vendor-fetch:
	dep ensure

fmt:
	gofmt -s -w $(GOFILES)

fmt-check:
	@diff=$$(gofmt -d $(GOFILES));               \
	if [ -n "$$diff" ]; then                     \
		echo "Please run 'make fmt' and commit"; \
		echo "$${diff}";                         \
		exit 1;                                  \
	fi

lint: PACKAGES = $(shell govendor list -no-status +local)
lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go get -u github.com/golang/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

test: PACKAGES = $(shell govendor list -no-status +local)
test: fmt-check
	echo "mode: count" > coverage-all.txt
	$(foreach pkg,$(PACKAGES),                                             \
		govendor test -coverprofile=coverage.txt -covermode=count $(pkg) || exit 1  \
		tail -n +2 coverage.txt >> coverage-all.txt;)
	go tool cover -html=coverage-all.txt

check: fmt-check test

bindata:
ifneq ($(GO_BINDATA_VERSION),$(REQ_GO_BINDATA_VERSION))
	go get -u github.com/shuLhan/go-bindata/...;
endif
	@go-bindata                                                           \
		-nometadata														  \
		-o common/bindata.go -pkg common 				                  \
	    common/data/...;
	@for tool in ${TOOL_LIST}; do                                         \
	  if [ -d $$tool/data ]; then                                         \
	    extra_dirs="$$tool/data/...";                                     \
	  fi;                                                                 \
	  go-bindata                                                          \
	  	-nometadata														  \
	    -o $$tool/$$tool/bindata.go -pkg $$tool                           \
	    common/data/... $$extra_dirs;                                     \
	done

build-all: bindata build-orchestrate build-registry

build-orchestrate:
	go build ${LDFLAGS} -o bin/ooni-orchestrate orchestrate/main.go
build-registry:
	go build ${LDFLAGS} -o bin/ooni-registry registry/main.go

orchestra: vendor build-all

orchestra-no-gitinfo: LDFLAGS = ${NOGI_LDFLAGS}
orchestra-no-gitinfo: vendor orchestra

release: fmt-check bindata
	go get github.com/goreleaser/goreleaser
	# XXX Not a fan of how it autogens the release notes, we should probably
	# extract them from our changelog and embed them using:
	# https://goreleaser.com/#releasing.custom_release_notes
	GITHUB_TOKEN=`cat .GITHUB_TOKEN` goreleaser --rm-dist

.PHONY: vendor build build-orchestrate build-registry release bindata build-all fmt fmt-check check
