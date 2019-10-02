FROM golang:latest

RUN mkdir /go/src/app

RUN go get -u github.com/golang/dep/cmd/dep

COPY . /go/src/app

WORKDIR /go/src/app

RUN dep ensure
RUN make build-orchestrate \
    && make build-registry

RUN cp bin/ooni-orchestrate /usr/bin/
RUN cp bin/ooni-registry /usr/bin/
