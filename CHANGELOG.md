# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)

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
