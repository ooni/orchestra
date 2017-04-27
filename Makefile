.PHONY: vendor build build-events build-notify build-registry

vendor:
	go get github.com/kardianos/govendor
	govendor sync proteus

build-events:
	go build -o bin/proteus-events proteus-events/main.go
build-notify:
	go build -o bin/proteus-notify proteus-notify/main.go
build-registry:
	go build -o bin/proteus-registry proteus-registry/main.go

build: build-events build-registry build-notify
