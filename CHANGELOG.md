# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)


### Hammer Beta-1

### Added
- Use the ruby 2.4 image [(#305)](https://github.com/ManageIQ/manageiq-pods/pull/305)
- Podify openidc hammer [(#296)](https://github.com/ManageIQ/manageiq-pods/pull/296)
- Backport enhancements made to master since Gaprindashvili [(#291)](https://github.com/ManageIQ/manageiq-pods/pull/291)

### Fixed
- Changed template to add miq-log config map [(#301)](https://github.com/ManageIQ/manageiq-pods/pull/301)
- Add the artemis username and password to the frontend pod [(#292)](https://github.com/ManageIQ/manageiq-pods/pull/292)

### Removed
- Remove artemis from the templates [(#294)](https://github.com/ManageIQ/manageiq-pods/pull/294)

## Gaprindashvili-3

### Added
- Add rolebindings for the orchestrator to the templates [(#271)](https://github.com/ManageIQ/manageiq-pods/pull/271)

## Gaprindashvili-2 - Released 2018-03-07

### Fixed
- Put the PG config map in the new new include location [(#267)](https://github.com/ManageIQ/manageiq-pods/pull/267)
- Mount the configmap into the directory the new image expects [(#264)](https://github.com/ManageIQ/manageiq-pods/pull/264)

## Gaprindashvili-1 - Released 2018-02-01

### Added
- Allow setting MIQ admin password during deployment [(#250)](https://github.com/ManageIQ/manageiq-pods/pull/250)
- Renaming auth_api to dbus_api service to reflect the new ManageIQ/dbus_api_service [(#230)](https://github.com/ManageIQ/manageiq-pods/pull/230)
- Rename API redirects httpd conf file [(#221)](https://github.com/ManageIQ/manageiq-pods/pull/221)
- Adding external authentication httpd configuration files [(#210)](https://github.com/ManageIQ/manageiq-pods/pull/210)
- Change wrap app/db PV definitions into kube templates [(#224)](https://github.com/ManageIQ/manageiq-pods/pull/224)
- Remove unnecessary service name environment variables [(#223)](https://github.com/ManageIQ/manageiq-pods/pull/223)
- Added support for the httpd auth-api service [(#204)](https://github.com/ManageIQ/manageiq-pods/pull/204)
- Enhance HTTPD pod liveness/readiness probes [(#218)](https://github.com/ManageIQ/manageiq-pods/pull/218)
- Support the httpd authentication configuration map. [(#201)](https://github.com/ManageIQ/manageiq-pods/pull/201)
- Drop all internal SSL [(#197)](https://github.com/ManageIQ/manageiq-pods/pull/197)
- Update to support foundational update of container-httpd for external authentication [(#194)](https://github.com/ManageIQ/manageiq-pods/pull/194)
- Allow MIQ database backup and restore via OpenShift jobs [(#189)](https://github.com/ManageIQ/manageiq-pods/pull/189)
- Increase PG Pod resource requirements [(#188)](https://github.com/ManageIQ/manageiq-pods/pull/188)
- Increase default httpd recreate strategy timeout [(#187)](https://github.com/ManageIQ/manageiq-pods/pull/187)
- Allow PG MIQ configuration overrides via configmap [(#185)](https://github.com/ManageIQ/manageiq-pods/pull/185)

### Fixed
- Create separate reverse proxying for websocket connections [(#208)](https://github.com/ManageIQ/manageiq-pods/pull/208)
- Add the ManageIQ copr repo to the app image [(#195)](https://github.com/ManageIQ/manageiq-pods/pull/195)
- Use the update:ui task rather than update:bower [(#190)](https://github.com/ManageIQ/manageiq-pods/pull/190)
- Do not set remote endpoint PORT on database-url on ext template [(#186)](https://github.com/ManageIQ/manageiq-pods/pull/186)
- Make websockets work again [(#255)](https://github.com/ManageIQ/manageiq-pods/pull/255)
