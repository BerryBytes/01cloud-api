# Release notes for v1.6.52


## Features

1. External Logger (Loki, Elastic Search, AWS S3 and Kafka) (#ZRO-3268) (#1245)

    - Support Loki, Elastic Search, AWS S3 and Kafka.
    - Prepare request and response schema.
    - Validation for external logger request input.

2.  Project activation and deativation implementation (#ZRO-3301) (#1246)

    - Design UsageHistory model schema.
    - Manage project usage history.
    - Prevent application and environment list if project status is deactivated (#ZRO-3309) (#1253).


## Enhancement

- Added validation for admin user while accessing resources (#1241)
- Added ErrorMessage into envoronment response doc model (#1247)
- Checked IsAdmin on all the get role cases (#1242)
- Tiggered notifiation message improvise (#ZRO-3292) (#1243)

## Bug or Regression

- Resolved send multiple email bugs  (#1244)
- Resloved projet activation bad gateway issue (#ZRO-3301) (#1248)
- Resolved create app on shared project if project owner balance amount available (#ZRO-3306) (#1251)
- Resolved cluster initial workflow log fetch issue (#ZRO-3281) (#1252)

## Dependencies

### Added
_Nothing has changed._

### Changed
_Nothing has changed._

### Removed
_Nothing has changed._
