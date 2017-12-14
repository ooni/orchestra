# proteus 0.2.0-rc.1 [2017-12-14]

Changes:

* Rename proteus-events to proteus-orchestrate

Adds:

* New orchestrate endpoints for /test-helpers, /collectors and /urls

# proteus 0.1.0-rc.1 [2017-09-29]

Changes:

* Improvements to testing and code style

* Add unittests that speak to a database

# proteus 0.1.0-beta.9 [2017-08-06]

* fix(registry): Language may be null
* fix(database): inconsistent migrations

# proteus 0.1.0-beta.8 [2017-08-02]

* refactor(Makefile): simplify the proteus targets dependencies
* feature(README.md): build and release instruction
* feature(events): fwd `click_action` for Android
* fix(registry): adjust syntax of `add_language_column`
* fix(Makefile): don't assume a tool has bindata
* regen(bindata): run `make bindata`
* fix(Makefile): always update embedded binary data

# proteus 0.1.0-beta.7 [2017-07-30]

* fix(proteus-events): do not propagate topic on Android

# proteus 0.1.0-beta.6 [2017-07-24]

* Improvements to the scheduling UI

* Add minimal progress bar

# proteus 0.1.0-beta.5 [2017-07-19]

* Fix bug in migration script

* Re-enable measurement job scheduling from UI

# proteus 0.1.0-beta.4 [2017-07-17]

* Fix bug in deletion of jobs

# proteus 0.1.0-beta.3 [2017-06-23]

* Add support for alert notifications

* Temporarily disable measurement job scheduling from UI

* Use gorush instead of proteus-notify

# proteus 0.1.0-beta.2 [2017-05-28]

* Fix the schema of proteus-events

# proteus 0.1.0-beta.1 [2017-05-26]

* Add support for Admins to manage jobs via the web UI

# proteus 0.1.0-dev [2017-05-13]

Development release of proteus.

Includes:

* Ability for probes to register with the orchestration registry
* Ability for probes to update the metadata about their probe
* Support for sending push notifications via Apple Push Notifcation service and Firebase Cloud Messagging
* Support for administrators to schedule measurements via a web interface
* Ability for probes to receive the tasks they have been notified about and mark them as accepted, rejected or done
* Multiple architecture build system

