# Changelog

## Unreleased

### Changed
- Added configurable repository merge style support (`default_merge_style`) in the repository resource.
- Standardized merge-style selection to the canonical API/UI values (no `rebase-ff` alias).
- Improved repository state reconciliation in create/read/update flows so merge-related fields stay known after apply.
- Added backward-compatible alias handling for organization identifiers:
  - Resource supports both `name` and `username`.
  - Data source supports both `name` and `org`.

### Default Merge Style Selection
Use the `default_merge_style` value that exactly matches the intended Gitea UI option:

- `merge` -> Selects "Create merge commit"
- `rebase` -> Selects "Rebase, then fast-forward"
- `rebase-merge` -> Selects "Rebase, then create merge commit"
- `squash` -> Selects "Create squash commit"
- `fast-forward-only` -> Selects "Fast-forward only"

Important:
- If you want "Rebase, then fast-forward" in the UI, set `default_merge_style = "rebase"`.
- Do not use `rebase-ff`; it is not a canonical selectable value.

### Fixed
- Fixed merge option plan/apply drift caused by unknown values being forwarded or restored incorrectly.
- Fixed team membership import test ID format (`team_id/username`).
- Fixed acceptance fixture mismatches for org/team/user resource inputs.

### Tests
- Added unit test coverage for:
  - `default_merge_style = fast-forward-only` forwarding.
  - full merge option forwarding in `buildEditRepoOption`.
- Added acceptance coverage for merge options set at create and changed post-create, including `fast-forward-only`.
- Updated org/repository/team/user acceptance tests to reflect current provider/server behavior and compatibility paths.
