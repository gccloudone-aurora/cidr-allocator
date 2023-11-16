# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
