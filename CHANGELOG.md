# Changelog

## Unreleased

### Changed
- Added configurable repository merge style support (`default_merge_style`) in the repository resource.
- Improved repository state reconciliation in create/read/update flows so merge-related fields stay known after apply.
- Added backward-compatible alias handling for organization identifiers:
  - Resource supports both `name` and `username`.
  - Data source supports both `name` and `org`.

### Fixed
- Fixed merge option plan/apply drift caused by unknown values being forwarded or restored incorrectly.
- Fixed team membership import test ID format (`team_id/username`).
- Fixed acceptance fixture mismatches for org/team/user resource inputs.

### Tests
- Added unit test coverage for:
  - `default_merge_style = rebase-ff` forwarding.
  - full merge option forwarding in `buildEditRepoOption`.
- Added acceptance coverage for merge options set at create and changed post-create.
- Updated org/repository/team/user acceptance tests to reflect current provider/server behavior and compatibility paths.
