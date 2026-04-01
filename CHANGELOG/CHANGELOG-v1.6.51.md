# Release notes for v1.6.51


## Feature

### External Secret (Hashicrop and GCP) (#1231)

- Support Hashicrop Vault Seceret.
- Support Google Cloud Platform (GCP) Secret.

### Notification, Email and Activity Log (#1236)

- Tiggred notification backup completed
- Send Email after backup completed
- Stored backup completd activity log

### Implement Changelog file (#1229)

- Changelogs for release full note.


## Enhancement

- Support external secret while cloned environment. (#1238)
- Total payment due and balance field added on dashboard model. (#1232)

## Bug or Regression

- Resolved external service login issue. (#1235)
- Resolved environment rollback issue. (#1233)
- Removed region from gcp credential model.(#1234)
- Added Variable field in project response model. (#1230)
- Removed vault credential from environment response.(#1237)

## Dependencies

### Added
_Nothing has changed._

### Changed
_Nothing has changed._

### Removed
_Nothing has changed._
