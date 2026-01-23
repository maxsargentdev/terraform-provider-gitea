# Terraform Provider Gitea - TODO

This document outlines potential additions to the provider based on available Gitea SDK functionality.

## High Priority - Core Repository Features

### 1. Repository Collaborators
- **Resource**: `gitea_repository_collaborator`
- **SDK Functions**: `AddCollaborator()`, `DeleteCollaborator()`, `CollaboratorPermission()`
- **Use case**: Manage direct user access to repositories with specific permissions (read, write, admin)
- **Priority**: HIGH - Essential for team workflows

### 2. Repository Labels
- **Resource**: `gitea_repository_label`
- **SDK Functions**: `CreateLabel()`, `EditLabel()`, `DeleteLabel()`, `GetRepoLabel()`
- **Use case**: Manage issue/PR labels for repositories
- **Priority**: HIGH - Issue/PR management

### 3. Repository Releases
- **Resource**: `gitea_repository_release`
- **SDK Functions**: `CreateRelease()`, `EditRelease()`, `DeleteRelease()`
- **Use case**: Manage software releases/tags with release notes
- **Priority**: HIGH - Common CI/CD use case

### 4. Repository Topics
- **Resource**: `gitea_repository_topics`
- **SDK Functions**: `SetRepoTopics()`, `AddRepoTopic()`, `DeleteRepoTopic()`
- **Use case**: Add searchable topics/tags to repositories
- **Priority**: HIGH - Improves discoverability

### 5. Tag Protection
- **Resource**: `gitea_repository_tag_protection`
- **SDK Functions**: `CreateTagProtection()`, `EditTagProtection()`, `DeleteTagProtection()`
- **SDK File**: `repo_tag_protection.go`
- **Use case**: Protect specific tags from being deleted or overwritten (similar to branch protection)
- **Priority**: HIGH - Complements existing branch protection

## Medium Priority - Repository Management

### 6. Repository Mirrors (Push)
- **Resource**: `gitea_repository_push_mirror`
- **SDK Functions**: `PushMirrors()`, `DeletePushMirror()`, `ListPushMirrors()`
- **SDK File**: `repo_mirror.go`
- **Use case**: Automatically push repository changes to remote mirrors
- **Priority**: MEDIUM

### 7. Repository Transfer
- **Resource**: `gitea_repository_transfer`
- **SDK Functions**: `TransferRepo()`, `AcceptRepoTransfer()`, `RejectRepoTransfer()`
- **SDK File**: `repo_transfer.go`
- **Use case**: Transfer repository ownership between users/organizations
- **Priority**: MEDIUM

### 8. Repository from Template
- **Resource**: `gitea_repository_from_template`
- **SDK Functions**: `CreateRepoFromTemplate()`
- **SDK File**: `repo_template.go`
- **Use case**: Create new repositories from template repositories
- **Priority**: MEDIUM

### 9. Repository Migration
- **Resource**: `gitea_repository_migrate`
- **SDK Functions**: `MigrateRepo()`
- **SDK File**: `repo_migrate.go`
- **Use case**: Import/migrate repositories from external sources (GitHub, GitLab, etc.)
- **Priority**: HIGH - Common onboarding scenario

### 10. Repository Branches
- **Resource**: `gitea_repository_branch`
- **SDK Functions**: `CreateBranch()`, `DeleteRepoBranch()`, `UpdateRepoBranch()`
- **SDK File**: `repo_branch.go`
- **Use case**: Programmatically create and manage branches
- **Priority**: MEDIUM

## Organization Features

### 11. Organization Labels
- **Resource**: `gitea_org_label`
- **SDK File**: `org_label.go`
- **Use case**: Manage org-wide labels that can be used across all org repositories
- **Priority**: MEDIUM

### 12. Organization Membership
- **Resource**: `gitea_org_membership`
- **SDK Functions**: `DeleteOrgMembership()`, `CheckOrgMembership()`
- **SDK File**: `org_member.go`
- **Use case**: Explicitly manage organization members (currently implicit through teams)
- **Priority**: MEDIUM

## User Features

### 13. User Email Management
- **Resource**: `gitea_user_email`
- **SDK Functions**: `AddEmail()`, `DeleteEmail()`, `ListEmails()`
- **SDK File**: `user_email.go`
- **Use case**: Manage additional email addresses for users
- **Priority**: LOW

### 14. User Settings
- **Resource**: `gitea_user_settings`
- **SDK Functions**: `UpdateUserSettings()`, `GetUserSettings()`
- **SDK File**: `user_settings.go`
- **Use case**: Configure user preferences and settings
- **Priority**: LOW

### 15. User Follow Management
- **Resource**: `gitea_user_follow` (or data source only)
- **SDK Functions**: `Follow()`, `Unfollow()`, `IsFollowing()`
- **SDK File**: `user_follow.go`
- **Use case**: Manage user following relationships
- **Priority**: LOW

## Issues & Pull Requests (Lower Priority - More Complex)

### 16. Issue Management
- **Resource**: `gitea_issue`
- **SDK File**: `issue.go`
- **Use case**: Create and manage issues programmatically
- **Priority**: LOW - Complex state management

### 17. Pull Requests
- **Resource**: `gitea_pull_request`
- **SDK Functions**: `CreatePullRequest()`, `EditPullRequest()`, `MergePullRequest()`
- **SDK File**: `pull.go`
- **Use case**: Automate PR creation and management
- **Priority**: LOW - Complex lifecycle management

### 18. Milestones
- **Resource**: `gitea_milestone`
- **SDK File**: `issue_milestone.go`
- **Use case**: Manage project milestones
- **Priority**: LOW

### 19. Issue/PR Comments
- **Resource**: `gitea_issue_comment`
- **SDK File**: `issue_comment.go`
- **Use case**: Manage comments on issues and pull requests
- **Priority**: LOW

## Commit Status Management

### 20. Commit Status
- **Resource**: `gitea_commit_status`
- **SDK Functions**: `CreateStatus()`, `ListStatuses()`, `GetCombinedStatus()`
- **SDK File**: `status.go`
- **Use case**: Set commit statuses for CI/CD integrations
- **Priority**: MEDIUM

## Package Management

### 21. Packages
- **Resource**: `gitea_package`
- **SDK Functions**: `GetPackage()`, `DeletePackage()`, `ListPackageFiles()`
- **SDK File**: `package.go`
- **Use case**: Manage packages in the Gitea package registry
- **Priority**: LOW

### 22. Package Cleanup Policies
- **Resource**: `gitea_package_cleanup_policy`
- **SDK Functions**: Not yet available in Gitea SDK
- **Use case**: Configure automatic cleanup/retention rules for package registry artifacts (e.g., keep last N versions, delete packages older than X days, regex-based cleanup)
- **Priority**: MEDIUM-HIGH
- **Status**: ⏳ **Blocked** - Awaiting Gitea API support
- **Notes**: 
  - Would enable automated artifact lifecycle management
  - Common requirement for CI/CD workflows to prevent storage bloat
  - Similar to GitHub Packages retention policies or Azure DevOps feed retention
  - Need to check if Gitea has added cleanup/retention APIs in recent versions
  - May need to contribute to Gitea SDK once API becomes available

## Data Sources to Add

### High Priority Data Sources
1. **`gitea_release`** - Read release information
2. **`gitea_releases`** - List all releases for a repository
3. **`gitea_collaborators`** - List repository collaborators
4. **`gitea_labels`** - List repository labels
5. **`gitea_tags`** - List repository tags
6. **`gitea_branches`** - List repository branches

### Medium Priority Data Sources
7. **`gitea_org_labels`** - List organization labels
8. **`gitea_org_members`** - List organization members
9. **`gitea_tag_protection`** - Read tag protection rules
10. **`gitea_collaborator`** - Read specific collaborator permissions
11. **`gitea_topics`** - List repository topics

### Lower Priority Data Sources
12. **`gitea_packages`** - List packages
13. **`gitea_issues`** - List repository issues
14. **`gitea_pull_requests`** - List repository pull requests
15. **`gitea_milestones`** - List repository milestones

## Most Impactful Additions (Recommended Implementation Order)

1. **Repository Collaborators** - Essential for team workflows, complements existing team management
2. **Repository Releases** - Common CI/CD use case, high demand feature
3. **Repository Migration** - Critical for onboarding new teams/projects
4. **Repository Labels** - Enables complete issue/PR workflow automation
5. **Tag Protection** - Natural extension of existing branch protection
6. **Repository Topics** - Simple to implement, improves repository organization
7. **Commit Status** - Important for CI/CD pipeline integrations
8. **Repository Branches** - Enables complete branch lifecycle management
9. **Repository Transfer** - Important for organizational restructuring
10. **Repository from Template** - Useful for standardization

## Known Provider Issues to Fix

Based on testing in `test/main.tf`:

1. **Repository Webhook** - Provider returns invalid timestamp values (`created_at`, `updated_at`)
2. **Repository Actions Secret** - Provider returns invalid timestamp values (`created_at`)
3. **Repository Actions Variable** - Potential timestamp issue (needs verification)
4. **Git Hooks** - Requires special permissions that may not be available in standard deployments

These timestamp issues need to be fixed in the respective resource implementations.

## Additional Considerations

### Admin Features
- The SDK includes admin functions (`admin_user.go`, `admin_org.go`, `admin_repo.go`) that could be exposed for server administration tasks

### Notifications
- The SDK includes notification management (`notifications.go`) that could be useful for automation

### Cron Jobs
- The SDK includes cron management (`admin_cron.go`) for scheduled tasks

### Settings
- Global settings management via `settings.go` could be useful for Gitea instance configuration
