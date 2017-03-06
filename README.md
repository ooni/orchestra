# Proteus

The OONI Probe Orchestration System.

This repository contains the various microservices that compose the OONI
Probe Orcehstration System.

**proteus-registry**

Status: ~COMPLETE

Is responsible for registering probes and keeping tabs on what their related
metadata is.

**proteus-notify**

Status: ~COMPLETE

Is responsible for dispatching notifications out to OPOS clients depending on
the capabilities they support.

**proteus-frontend**

Status: WIP

Is the frontend to:

* Setup schedules

* View active schedules

* View active probes

**proteus-events**

Status: Not Started

Is responsible for receiving events via the admin interface and triggering
notifications via **proteus-notify**.

Can also be used to view the event history.


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
**Path**: `/api/v1/clients`
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
**Path**: `/api/v1/clients/$CLIENT_ID`
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
