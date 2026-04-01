# Release notes for v1.6.50


## Feature

### External Secret

- External Secret supported for templete and non-templete
- External Secret re-sync, retry and logs

### Middleware Log Level

- Added Debug log level support

## Enhancement

- Dockerfile updated

## Bug or Regression

- Resolved organization count issue(#1220)
- Resolved subscription active true issue (#1221)
- Resolved custom domain value set interrupt user variables value (#1222)
- Resolved preload on project and organization models. (#1223)

## Other (Cleanup or Flake)

- Removed log level INFO parts.
- Removed send mail funcationalites and migrate it to notification microservice.

## Dependencies

### Added
_Nothing has changed._

### Changed
- 	github.com/joho/godotenv: [v1.3.0 → v1.4.0](https://github.com/joho/godotenv/compare/v1.3.0...v1.4.0)
- github.com/golang/protobuf: [v1.4.2 → v1.5.0](https://github.com/golang/protobuf/compare/v1.4.2...v1.5.0)
- k8s.io/api v0.17.2: [v0.17.2 → v0.18.3]

### Removed
- cloud.google.com/go/datastore v1.1.0
- cloud.google.com/go/firestore v1.2.0
- github.com/matcornic/hermes/v2 v2.1.0
- github.com/twinj/uuid v1.0.0
