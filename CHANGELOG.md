# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Added compatibility with legacy GestioIP 3.2 deployments that protect both API and frontend flows with Basic Auth.
- Added automated acceptance tests for GestioIP 3.2 and GestioIP 3.5.
- Added a manual GitHub Actions workflow for running acceptance tests against lab environments.

### Changed

- Network reads now fall back to frontend HTML parsing when legacy installations do not expose a usable JSON response for `listNetworks`.
- VLAN import now preserves existing remote values when optional attributes are not explicitly configured.

## [0.2.3] - 2026-03-23

### Added

- Added Terraform Registry provider index documentation.
- Added public testing disclaimers to README and provider docs.

## [0.2.2] - 2026-03-23

### Fixed

- Fixed non-interactive GPG signing in GitHub Actions releases.

## [0.2.1] - 2026-03-23

### Added

- Added Terraform Registry publishing assets and release automation.

## [0.2] - 2026-03-19

### Added

- Added `allow_overwrite` provider configuration for `host`, `network`, and `vlan`.
- Added `terraform import` support for `host`, `network`, and `vlan`.
- Added provider and resource documentation in `README.md` and `docs/`.

### Changed

- Documented overwrite and import behavior for existing GestioIP objects.

## [0.1] - 2026-03-18

### Added

- Initial provider scaffolding using Terraform Plugin Framework.
- Added `gestioip_network`, `gestioip_host`, and `gestioip_vlan` resources and data sources.
- Added hybrid GestioIP free-edition integration using `intapi.cgi` and frontend CGI flows.
