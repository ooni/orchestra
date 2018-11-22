# OONI Orchestra

[![Build Status](https://travis-ci.org/ooni/orchestra.svg)](https://travis-ci.org/ooni/orchestra)
[![Release](https://img.shields.io/github/release/ooni/orchestra.svg?style=flat-square)](https://github.com/ooni/orchestra/releases)


The OONI Probe Orchestration System.

This repository contains the various microservices that compose the OONI
Probe Orcehstration System.

## Getting started

You should have a working golang installation. Then you can:

```
go get github.com/ooni/orchestra
```

You will also need to install:

```
go get -u github.com/shuLhan/go-bindata/...
```

To build all the binaries, run:

```
make orchestra
```

You should now have inside of `./bin` a series of binaries. The ones you care
about are:

```
./bin/ooni-orchestra
./bin/ooni-registry
```

They both take a config file--you can use
`orchestrate/ooni-orchestrate.toml.example` (or the `registry` one) as
examples, respectively.

To aid local testing you should set up a local postgres instance with a user
and a database by running:
```
CREATE USER proteus WITH PASSWORD 'changeme';
CREATE DATABASE proteus;
GRANT ALL PRIVILEGES ON DATABASE proteus to proteus;
```

You should then change the config line in the `database` section for the key
`url` to something like:

```
[database]
url = "postgres://proteus:changeme@127.0.0.1:32768/proteus?sslmode=disable"
```

For both `orchestrate/ooni-orchestrate.toml` and
`registry/ooni-registry.toml`.

You should then be able to start the services by running:

```
./bin/ooni-orchestrate --config orchestrate/ooni-orchestrate.toml start
```


## Components

**registry**

Is responsible for registering probes and keeping tabs on what their related
metadata is.

**frontend**

Is the frontend to:

* Setup schedules

* View active schedules

* View active probes

**orchestrate**

Is responsible for receiving events via the admin interface and triggering
notifications via `gorush`.

Can also be used to view the event history.

## Building and releasing

- Make sure the `GOPATH` environment variable is set. (For reference, the setup
  of sbs is `export GOPATH=$HOME` with repositories in `$HOME/src/`; e.g. this
  repository is `$HOME/src/github.com/ooni/orchestra`).

- Of course, you also need to have golang installed.

- To build for development, run `make orchestra`.

- To create a release, run `make release`.

Checklist before tagging a release:

- Make sure you have updated the changelog

- Make sure you bumped the version number in:

    - `Makefile`

    - `common/orchestra_info.go`

- Make sure all unittests are passing (`make check`)

## Notifications specification

A client needs to register to the proteus-registry service.

The canonical address for it shall be `https://registry.XXX.YYY.ZZZ/`. We
also support the cloudfronted domain as following.

### Registration

Upon first running the app the client needs to registry with the notification
service.

Once a client is registered they can update the various metadata related to the probe by means of an update request (detailed in the following section).

To register a probe a the following HTTPS request is issued:

**Method**: `POST`
**Path**: `/api/v1/register`
**Body**:
```json
{
  "probe_cc": "IT",
  "probe_asn": "AS0",
  "platform": "android",
  "software_name": "ooniprobe-android",
  "software_version": "0.1.1",
  "software_language": "IT",
  "supported_tests": ["tcp_connect", "web_connectivity"],
  "network_type": "wifi",
  "available_bandwidth": "100",
  "token": "TOKEN_ID"
}
```

In particular the `token` field represents the Device Token.

The registration service will return a `client_id` to be used to update the Device Token as well as other metadata.

The response looks like this:

```json
{"client_id": "XXX-YYY-ZZZ-TTT"}
```

### Update metadata

In order do update the metadata you to issue the following request:

**Method**: `PUT`
**Path**: `/api/v1/update/$CLIENT_ID`
**Body**:
```json
{
  "probe_cc": "IT",
  "probe_asn": "AS0",
  "platform": "android",
  "software_name": "ooniprobe-android",
  "software_version": "0.1.1",
  "supported_tests": ["tcp_connect", "web_connectivity"],
  "network_type": "wifi",
  "available_bandwidth": "100",
  "token": "NEW_TOKEN_ID"
}
```

The server will respond with:

```json
{"status": "ok"}
```

in case of an error:

```json
{"error": "ERROR_MESSAGE"}
```

### Notifications

A notification includes in the payload of the silent notification a pingback
URL that the client needs to connect to in order to receive the task that it
need to run to perform the actual measurement.
