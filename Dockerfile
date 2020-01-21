FROM golang:latest
RUN mkdir /go/src/app
COPY . /go/src/app
WORKDIR /go/src/app
RUN make build-orchestrate && make build-registry
RUN cp bin/ooni-orchestrate /usr/bin/
RUN cp bin/ooni-registry /usr/bin/
