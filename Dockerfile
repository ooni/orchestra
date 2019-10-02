FROM golang:latest

WORKDIR /app

RUN go get -u github.com/golang/dep/cmd/dep

COPY . .

RUN dep ensure
RUN make build-all

COPY bin/ooni-orchestrate /usr/bin/
COPY bin/ooni-registry /usr/bin/
