# Proteus

The OONI Probe Orchestration System.

This repository contains the various microservices that compose the OONI
Probe Orcehstration System.

**proteus-registry**

Status: ~COMPLETE

Is responsible for registering probes and keeping tabs on what their related
metadata is.

**proteus-notify**

Status: WIP

Is responsible for dispatching notifications out to OPOS clients depending on
the capabilities they support.

**proteus-events**

Status: Not started

Is responsible for receiving events via the admin interface and triggering
notifications via **proteus-notify**.

Can also be used to view the event history

