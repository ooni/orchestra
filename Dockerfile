FROM golang:latest

RUN mkdir /go/src/app

ARG GO111MODULE=off
RUN go get -u -v github.com/golang/dep/cmd/dep
RUN go get -u -v github.com/shuLhan/go-bindata/cmd/go-bindata

COPY . /go/src/app

WORKDIR /go/src/app

RUN dep ensure
RUN make build-orchestrate \
    && make build-registry

RUN cp bin/ooni-orchestrate /usr/bin/
RUN cp bin/ooni-registry /usr/bin/
