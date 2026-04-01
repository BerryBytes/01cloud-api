# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> Note: Historical changelog entries prior to version 1.6.58 can be found in the `CHANGELOG/` directory.

## [1.6.58] - 2024-03-27

### Added
- Implemented project validation API

### Changed
- Updated base URL references
- Updated file path for configs to `/data/public`

### Fixed
- Fixed cronjob create issue
- Fixed routing override and refactored related logic
- Fixed Cronjob logs retrieval
- Resolved nil pointer dereference on `injectSecretPatcherEnvs`
- Fixed access modes (now set to RWO or RWX, avoiding 'none')
- Fixed download kubeconfig file functionality
- Resolved vcluster validate API endpoint
- Addressed nil pointer for env cluster checks
