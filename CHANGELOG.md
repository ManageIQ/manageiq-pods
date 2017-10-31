# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)


## Unreleased as of Sprint 72 ending 2017-10-30

### Added
- Enhance app-frontend probes [(#239)](https://github.com/ManageIQ/manageiq-pods/pull/239)

## Unreleased as of Sprint 71 ending 2017-10-16

### Added
- Renaming auth_api to dbus_api service to reflect the new ManageIQ/dbus_api_service [(#230)](https://github.com/ManageIQ/manageiq-pods/pull/230)
- Rename API redirects httpd conf file [(#221)](https://github.com/ManageIQ/manageiq-pods/pull/221)
- Adding external authentication httpd configuration files [(#210)](https://github.com/ManageIQ/manageiq-pods/pull/210)

## Unreleased as of Sprint 70 ending 2017-10-02

### Added
- Change wrap app/db PV definitions into kube templates [(#224)](https://github.com/ManageIQ/manageiq-pods/pull/224)
- Remove unnecessary service name environment variables [(#223)](https://github.com/ManageIQ/manageiq-pods/pull/223)
- Added support for the httpd auth-api service [(#204)](https://github.com/ManageIQ/manageiq-pods/pull/204)

## Unreleased as of Sprint 69 ending 2017-09-18

### Added
- Enhance HTTPD pod liveness/readiness probes [(#218)](https://github.com/ManageIQ/manageiq-pods/pull/218)

## Unreleased as of Sprint 68 ending 2017-09-04

### Added
- Authentication
  - Support the httpd authentication configuration map. [(#201)](https://github.com/ManageIQ/manageiq-pods/pull/201)

### Fixed
- Platform
  - Create separate reverse proxying for websocket connections [(#208)](https://github.com/ManageIQ/manageiq-pods/pull/208)

## Unreleased as of Sprint 67 ending 2017-08-21

### Added
- Platform
  - Drop all internal SSL [(#197)](https://github.com/ManageIQ/manageiq-pods/pull/197)
  - Update to support foundational update of container-httpd for external authentication [(#194)](https://github.com/ManageIQ/manageiq-pods/pull/194)
- Providers
  - Allow MIQ database backup and restore via OpenShift jobs [(#189)](https://github.com/ManageIQ/manageiq-pods/pull/189)

### Fixed
- Build
  - Add the ManageIQ copr repo to the app image [(#195)](https://github.com/ManageIQ/manageiq-pods/pull/195)

## Unreleased as of Sprint 66 ending 2017-08-07

### Added
- Providers
  - Increase PG Pod resource requirements [(#188)](https://github.com/ManageIQ/manageiq-pods/pull/188)
- Platform
  - Increase default httpd recreate strategy timeout [(#187)](https://github.com/ManageIQ/manageiq-pods/pull/187)
  - Allow PG MIQ configuration overrides via configmap [(#185)](https://github.com/ManageIQ/manageiq-pods/pull/185)

### Fixed
- Platform
  - Use the update:ui task rather than update:bower [(#190)](https://github.com/ManageIQ/manageiq-pods/pull/190)
  - Do not set remote endpoint PORT on database-url on ext template [(#186)](https://github.com/ManageIQ/manageiq-pods/pull/186)
