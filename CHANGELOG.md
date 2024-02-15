# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2024-02-13
### Breaking Change
- feat(v1.0.0): initial release created and all previous versions migrated to Kubebuilder GO/v4
- update(webhooks): removed default and validating webhooks
  - webhooks were removed since the use-case for this operator primarily requires configuration prior to networking being available. As a result, there was little benefit added with the pre-existing webhooks implementation and it created a lot of additional complexity
- feat(helm): added helm chart to source repository

## [v0.4.4] - 2023-12-21
### Changed
- refactor(main): leader-election ID (name) made to be more descriptive and align with best practices
- refactor(main): simplify configuration model for the controller-manager

## [v0.4.3] - 2023-11-16
### Fixed
- fix(dockerfile): put arg above use to work with docker buildx

## [v0.4.2] - 2023-11-16
### Changed
- ci(github): added GitHub workflows
- ci(dockerfile): parametarize image repository
- ci(gitlab): pass build args to use artifactory as image base and target
- docs(README): documentation for installation/usage and design/architecture

## [v0.4.1] - 2023-11-16

### Fixed
- Event rejection from Kube API server resolved by adding appropriate RBAC
### Added
- missing unit tests

## [v0.4.0] - 2023-03-06
### Added
- Prometheus metrics support

## [v0.3.0] - 2023-02-22
### Added
- cluster-wide node CIDR collision detection and avoidance

## [v0.2.1] - 2023-02-16
### Fixed
- issue where only a single nodeSelector was evaluated fixed

## [v0.2.0] - 2023-02-07
### Added
- implemented NodeCIDRAllocation resource Health and Status

## [v0.1.1] - 2023-02-02
### Changed
- Cleaned up manifests
- refactored to align with golang style guide

## [v0.1.0] - 2023-02-01
### Added
- Initial Release
