# # ============================================================================
# # Comprehensive Team Resource Testing
# # ============================================================================
# # This file creates a wide variety of teams to test different configurations
# # and edge cases for the gitea_team resource.

# # ----------------------------------------------------------------------------
# # Basic Teams - Minimal Configuration
# # ----------------------------------------------------------------------------

# resource "gitea_team" "minimal" {
#   org  = gitea_org.test_org.username
#   name = "minimal-team"
  
#   units_map = {
#     "repo.code" = "read"
#   }
# }

# resource "gitea_team" "basic_write" {
#   org  = gitea_org.test_org.username
#   name = "basic-writers"
  
#   units_map = {
#     "repo.code" = "write"
#   }
  
#   description = "Basic team with write access"
# }

# # ----------------------------------------------------------------------------
# # Permission Variations - Different Access Levels
# # ----------------------------------------------------------------------------

# resource "gitea_team" "readers_only" {
#   org         = gitea_org.test_org.username
#   name        = "readers-only"
#   description = "Team with read-only access to code and issues"
  
#   units_map = {
#     "repo.code"   = "read"
#     "repo.issues" = "read"
#   }
# }

# resource "gitea_team" "code_admins" {
#   org         = gitea_org.test_org.username
#   name        = "code-admins"
#   description = "Team with maximum (write) access to code repositories"
  
#   units_map = {
#     "repo.code" = "write"
#   }
  
#   can_create_org_repo = true
# }

# resource "gitea_team" "mixed_permissions" {
#   org         = gitea_org.test_org.username
#   name        = "mixed-permissions"
#   description = "Team with mixed permission levels across different units"
  
#   units_map = {
#     "repo.code"     = "write"
#     "repo.issues"   = "write"
#     "repo.pulls"    = "write"
#     "repo.releases" = "read"
#   }
# }

# # ----------------------------------------------------------------------------
# # Full Access Teams - All Units Configured
# # ----------------------------------------------------------------------------

# resource "gitea_team" "full_write_access" {
#   org         = gitea_org.test_org.username
#   name        = "full-write-team"
#   description = "Team with write access to all repository units"
  
#   can_create_org_repo = false
  
#   units_map = {
#     "repo.code"       = "write"
#     "repo.issues"     = "write"
#     "repo.pulls"      = "write"
#     "repo.releases"   = "write"
#     "repo.wiki"       = "write"
#     "repo.ext_wiki"   = "write"
#     "repo.ext_issues" = "write"
#     "repo.projects"   = "write"
#     "repo.packages"   = "write"
#     "repo.actions"    = "write"
#   }
# }

# resource "gitea_team" "full_admin_access" {
#   org         = gitea_org.test_org.username
#   name        = "full-admin-team"
#   description = "Team with maximum (write) access to all repository units"
  
#   can_create_org_repo = true
  
#   units_map = {
#     "repo.code"       = "write"
#     "repo.issues"     = "write"
#     "repo.pulls"      = "write"
#     "repo.releases"   = "write"
#     "repo.wiki"       = "write"
#     "repo.ext_wiki"   = "write"
#     "repo.ext_issues" = "write"
#     "repo.projects"   = "write"
#     "repo.packages"   = "write"
#     "repo.actions"    = "write"
#   }
# }

# resource "gitea_team" "granular_access" {
#   org         = gitea_org.test_org.username
#   name        = "granular-access"
#   description = "Team with carefully controlled granular permissions"
  
#   units_map = {
#     "repo.code"       = "write"
#     "repo.issues"     = "write"
#     "repo.pulls"      = "write"
#     "repo.releases"   = "read"
#     "repo.wiki"       = "none"
#     "repo.ext_wiki"   = "none"
#     "repo.ext_issues" = "read"
#   }
# }

# # ----------------------------------------------------------------------------
# # Special Configuration Teams
# # ----------------------------------------------------------------------------

# resource "gitea_team" "repo_creators" {
#   org         = gitea_org.test_org.username
#   name        = "repo-creators"
#   description = "Team allowed to create organization repositories"
  
#   can_create_org_repo = true
  
#   units_map = {
#     "repo.code"   = "write"
#     "repo.issues" = "write"
#     "repo.pulls"  = "write"
#   }
# }

# resource "gitea_team" "all_repos_access" {
#   org         = gitea_org.test_org.username
#   name        = "all-repos-team"
#   description = "Team with automatic access to all repositories (WARNING: see docs)"
  
#   includes_all_repositories = true
  
#   units_map = {
#     "repo.code"   = "read"
#     "repo.issues" = "read"
#   }
# }

# # ----------------------------------------------------------------------------
# # Specialized Teams - Specific Workflows
# # ----------------------------------------------------------------------------

# resource "gitea_team" "issue_managers" {
#   org         = gitea_org.test_org.username
#   name        = "issue-managers"
#   description = "Team focused on issue management"
  
#   units_map = {
#     "repo.code"   = "read"
#     "repo.issues" = "write"
#   }
# }

# resource "gitea_team" "release_managers" {
#   org         = gitea_org.test_org.username
#   name        = "release-managers"
#   description = "Team responsible for managing releases"
  
#   units_map = {
#     "repo.code"     = "read"
#     "repo.releases" = "write"
#   }
# }

# resource "gitea_team" "pr_reviewers" {
#   org         = gitea_org.test_org.username
#   name        = "pr-reviewers"
#   description = "Team dedicated to reviewing pull requests"
  
#   units_map = {
#     "repo.code"  = "read"
#     "repo.pulls" = "write"
#   }
# }

# resource "gitea_team" "wiki_editors" {
#   org         = gitea_org.test_org.username
#   name        = "wiki-editors"
#   description = "Team focused on documentation via wiki"
  
#   units_map = {
#     "repo.code" = "read"
#     "repo.wiki" = "write"
#   }
# }

# resource "gitea_team" "external_integrations" {
#   org         = gitea_org.test_org.username
#   name        = "external-integrations"
#   description = "Team managing external integrations"
  
#   units_map = {
#     "repo.code"       = "read"
#     "repo.ext_wiki"   = "write"
#     "repo.ext_issues" = "write"
#   }
# }

# resource "gitea_team" "project_managers" {
#   org         = gitea_org.test_org.username
#   name        = "project-managers"
#   description = "Team managing project boards"
  
#   units_map = {
#     "repo.code"     = "read"
#     "repo.projects" = "write"
#   }
# }

# resource "gitea_team" "package_publishers" {
#   org         = gitea_org.test_org.username
#   name        = "package-publishers"
#   description = "Team responsible for publishing packages"
  
#   units_map = {
#     "repo.code"     = "read"
#     "repo.packages" = "write"
#   }
# }

# resource "gitea_team" "ci_cd_managers" {
#   org         = gitea_org.test_org.username
#   name        = "ci-cd-managers"
#   description = "Team managing CI/CD actions and workflows"
  
#   units_map = {
#     "repo.code"    = "read"
#     "repo.actions" = "write"
#   }
# }

# # ----------------------------------------------------------------------------
# # Edge Cases and No-Access Teams
# # ----------------------------------------------------------------------------

# resource "gitea_team" "no_direct_access" {
#   org         = gitea_org.test_org.username
#   name        = "no-direct-access"
#   description = "Team with no repository unit permissions"
  
#   units_map = {
#     "repo.code"       = "none"
#     "repo.issues"     = "none"
#     "repo.pulls"      = "none"
#     "repo.releases"   = "none"
#     "repo.wiki"       = "none"
#     "repo.ext_wiki"   = "none"
#     "repo.ext_issues" = "none"
#     "repo.projects"   = "none"
#     "repo.packages"   = "none"
#     "repo.actions"    = "none"
#   }
# }

# resource "gitea_team" "code_only_write" {
#   org         = gitea_org.test_org.username
#   name        = "code-only-write"
#   description = "Team with write access only to code, nothing else"
  
#   can_create_org_repo = true
  
#   units_map = {
#     "repo.code"       = "write"
#     "repo.issues"     = "none"
#     "repo.pulls"      = "none"
#     "repo.releases"   = "none"
#     "repo.wiki"       = "none"
#     "repo.ext_wiki"   = "none"
#     "repo.ext_issues" = "none"
#     "repo.projects"   = "none"
#     "repo.packages"   = "none"
#     "repo.actions"    = "none"
#   }
# }
