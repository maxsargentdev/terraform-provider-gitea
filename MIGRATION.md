# Migrating from go-gitea/gitea Provider

This guide helps you migrate from the [existing Gitea Terraform provider](https://registry.terraform.io/providers/go-gitea/gitea/latest) (built with the old SDK) to this provider (built with the new Terraform Plugin Framework).

## Why Migrate?

This provider offers several improvements:

- **New Terraform Plugin Framework**: Better performance, improved type safety, and modern architecture
- **Enhanced Team Permissions**: Full support for Gitea's `units_map` API for granular repository permissions
- **Active Development**: More frequent updates and feature additions
- **Better Error Handling**: Improved retry logic for concurrent operations
- **No Drift Issues**: Fixes persistent drift problems with team permissions

## Breaking Changes

### 1. Team Resource - Units Map

The `gitea_team` resource has a **different API** for permissions:

**Old Provider (go-gitea/gitea):**
```hcl
resource "gitea_team" "example" {
  organisation = "myorg"
  name         = "developers"
  permission   = "write"  # Simple permission level
}
```

**New Provider (this one):**
```hcl
resource "gitea_team" "example" {
  org  = "myorg"  # Note: 'org' instead of 'organisation'
  name = "developers"
  
  # Granular per-unit permissions
  units_map = {
    "repo.code"       = "write"
    "repo.issues"     = "write"
    "repo.pulls"      = "write"
    "repo.releases"   = "none"
    "repo.ext_wiki"   = "none"
    "repo.ext_issues" = "read"
    "repo.actions"    = "write"
  }
}
```

### 2. Team Repository Assignment

**Old Provider:**
```hcl
resource "gitea_team" "example" {
  organisation = "myorg"
  name         = "developers"
  permission   = "write"
  
  # Repositories as a list
  repositories = [
    "repo1",
    "repo2"
  ]
}
```

**New Provider:**
```hcl
resource "gitea_team" "example" {
  org  = "myorg"
  name = "developers"
  
  units_map = {
    "repo.code" = "write"
    # ... other permissions
  }
}

# Separate resource for each repository assignment
resource "gitea_team_repository" "example_repo1" {
  org             = "myorg"
  team_name       = gitea_team.example.name
  repository_name = "repo1"
}

resource "gitea_team_repository" "example_repo2" {
  org             = "myorg"
  team_name       = gitea_team.example.name
  repository_name = "repo2"
}
```

### 3. Provider Configuration

Update your provider configuration:

**Old:**
```hcl
terraform {
  required_providers {
    gitea = {
      source  = "go-gitea/gitea"
      version = "~> 0.x"
    }
  }
}

provider "gitea" {
  base_url = "https://gitea.example.com"
  token    = var.gitea_token
}
```

**New:**
```hcl
terraform {
  required_providers {
    gitea = {
      source  = "maxsargendev/gitea"  # Update namespace
      version = "~> 1.0"
    }
  }
}

provider "gitea" {
  hostname = "https://gitea.example.com"  # Note: 'hostname' instead of 'base_url'
  username = "admin"                       # May require username
  password = var.gitea_password            # Or token
  # token    = var.gitea_token             # Alternative to username/password
}
```

## Migration Strategy

### Option 1: Import Existing Resources (Recommended)

This approach imports your existing Gitea resources into the new provider without destroying/recreating them.

#### Step 1: Update Provider Configuration

```bash
# Update your provider version in terraform block
terraform init -upgrade
```

#### Step 2: Remove Old State

```bash
# For each resource, remove from old provider state
terraform state rm gitea_team.example
terraform state rm gitea_repository.example
# ... etc
```

#### Step 3: Import Into New Provider

All resources in this provider support `terraform import`:

**Teams:**
```bash
# Import by team ID (numeric)
terraform import gitea_team.example 123
```

**Repositories:**
```bash
# Import by owner/repo_name
terraform import gitea_repository.example myorg/myrepo
```

**Team Memberships:**
```bash
# Import by team_id/username
terraform import gitea_team_membership.example 123/johndoe
```

**Team Repository Assignments:**
```bash
# Import by org/team_name/repository_name
terraform import gitea_team_repository.example myorg/developers/myrepo
```

**Organizations:**
```bash
# Import by org name
terraform import gitea_org.example myorg
```

**Users:**
```bash
# Import by username
terraform import gitea_user.example johndoe
```

**Other Resources:**
- See each resource's documentation for import format
- All resources include import instructions in their docs

#### Step 4: Update Resource Definitions

Update your `.tf` files to match the new provider's schema, particularly:
- Change `organisation` to `org`
- Convert team `permission` to `units_map`
- Split team `repositories` into separate `gitea_team_repository` resources

#### Step 5: Plan and Verify

```bash
terraform plan
```

Review the plan carefully. You should see minimal or no changes if you've correctly matched the schema.

### Option 2: Separate Workspace Testing

If you want to test the new provider before migrating:

#### Step 1: Create a Separate Testing Workspace

```bash
# Copy your existing configuration to a test directory
cp -r terraform/ terraform-new-provider-test/
cd terraform-new-provider-test/
```

#### Step 2: Update Provider in Test Workspace

```hcl
terraform {
  required_providers {
    gitea = {
      source  = "maxsargendev/gitea"  # New provider
      version = "~> 1.0"
    }
  }
}
```

#### Step 3: Test with Non-Production Resources

Create test resources to verify the new provider works:

```bash
terraform init
terraform plan
terraform apply
```

#### Step 4: Once Validated, Migrate Production

After successful testing, use Option 1 to migrate your production workspace.

**Note**: You cannot run both providers simultaneously in the same workspace since they both use the `gitea` name. You must fully switch from one to the other.

### Option 3: Fresh Start (If Acceptable)

If your Gitea infrastructure can be recreated:

#### Step 1: Export Current State

Document your current setup or export resources from Gitea.

#### Step 2: Destroy Old Resources

```bash
terraform destroy
```

#### Step 3: Update Provider

Update your Terraform configuration to use the new provider.

#### Step 4: Apply New Configuration

```bash
terraform init -upgrade
terraform apply
```

## Resource Mapping

| Old Provider | New Provider | Import Format | Notes |
|--------------|--------------|---------------|-------|
| `gitea_team` | `gitea_team` | `<team_id>` | Schema changed: use `units_map` |
| `gitea_repository` | `gitea_repository` | `<owner>/<repo>` | Similar schema |
| `gitea_org` | `gitea_org` | `<org_name>` | Similar schema |
| `gitea_user` | `gitea_user` | `<username>` | Similar schema |
| N/A | `gitea_team_repository` | `<org>/<team>/<repo>` | **New resource** for repo assignments |
| N/A | `gitea_team_membership` | `<team_id>/<username>` | May have existed in old provider |
| `gitea_oauth2_app` | `gitea_oauth2_app` | Check docs | Similar schema |
| `gitea_public_key` | `gitea_public_key` | Check docs | Similar schema |
| `gitea_repository_key` | `gitea_repository_key` | Check docs | Similar schema |

## Common Pitfalls

### 1. Team Permissions Mapping

When converting `permission` to `units_map`:

- `read` → All units set to `"read"`
- `write` → Code units set to `"write"`, others as needed
- `admin` → All units set to `"admin"`

**Helper Script:**
```python
# Convert simple permission to units_map
def convert_permission(permission):
    if permission == "read":
        return {unit: "read" for unit in [
            "repo.code", "repo.issues", "repo.pulls", 
            "repo.releases", "repo.wiki", "repo.actions"
        ]}
    elif permission == "write":
        return {unit: "write" for unit in [
            "repo.code", "repo.issues", "repo.pulls", "repo.actions"
        ]}
    elif permission == "admin":
        return {unit: "admin" for unit in [
            "repo.code", "repo.issues", "repo.pulls", 
            "repo.releases", "repo.wiki", "repo.actions"
        ]}
```

### 2. Repository Lists to Individual Resources

Use `for_each` to convert lists:

```hcl
locals {
  team_repos = toset(["repo1", "repo2", "repo3"])
}

resource "gitea_team_repository" "assignments" {
  for_each = local.team_repos
  
  org             = "myorg"
  team_name       = "developers"
  repository_name = each.value
}
```

### 3. Provider Authentication

The new provider may require different auth:
- Check if username/password is required instead of token
- Verify hostname format (with or without `/api/v1`)

## Testing Your Migration

1. **Start Small**: Migrate one team or repository first
2. **Use `terraform plan`**: Verify no unexpected changes
3. **Test Imports**: Ensure import works before removing old state
4. **Backup State**: Always backup `terraform.tfstate` before migration
5. **Validate**: Run `terraform validate` after updating configuration

## Rollback Plan

If you need to rollback:

1. **Restore State Backup**: 
   ```bash
   cp terraform.tfstate.backup terraform.tfstate
   ```

2. **Downgrade Provider**:
   ```bash
   terraform init -upgrade
   ```

3. **Revert Configuration**: Use version control to restore old `.tf` files

## Getting Help

- **Issues**: Check the [GitHub issues](https://github.com/maxsargendev/terraform-provider-gitea/issues)
- **Documentation**: Review the `/docs` directory for each resource
- **Examples**: See `/examples` directory for working configurations
- **Import Format**: Every resource doc includes import instructions

## Migration Checklist

- [ ] Backup current Terraform state
- [ ] Review breaking changes (especially team resources)
- [ ] Update provider configuration
- [ ] Test import on a single resource
- [ ] Update resource schemas in `.tf` files
- [ ] Convert team repositories to separate resources
- [ ] Run `terraform plan` and review
- [ ] Apply changes incrementally
- [ ] Verify resources in Gitea UI
- [ ] Update CI/CD pipelines with new provider version
- [ ] Document any custom workarounds

## Example: Complete Team Migration

**Before (old provider):**
```hcl
resource "gitea_team" "developers" {
  organisation = "acme"
  name         = "developers"
  permission   = "write"
  
  repositories = [
    "backend",
    "frontend",
    "docs"
  ]
  
  members = [
    "alice",
    "bob"
  ]
}
```

**After (new provider):**
```hcl
# Step 1: Create team with granular permissions
resource "gitea_team" "developers" {
  org         = "acme"
  name        = "developers"
  description = "Development team"
  
  units_map = {
    "repo.code"       = "write"
    "repo.issues"     = "write"
    "repo.pulls"      = "write"
    "repo.releases"   = "read"
    "repo.ext_wiki"   = "none"
    "repo.ext_issues" = "none"
    "repo.actions"    = "write"
  }
}

# Step 2: Assign repositories individually
resource "gitea_team_repository" "dev_backend" {
  org             = "acme"
  team_name       = gitea_team.developers.name
  repository_name = "backend"
}

resource "gitea_team_repository" "dev_frontend" {
  org             = "acme"
  team_name       = gitea_team.developers.name
  repository_name = "frontend"
}

resource "gitea_team_repository" "dev_docs" {
  org             = "acme"
  team_name       = gitea_team.developers.name
  repository_name = "docs"
}

# Step 3: Assign members individually (if resource exists)
resource "gitea_team_membership" "dev_alice" {
  team_id  = gitea_team.developers.id
  username = "alice"
}

resource "gitea_team_membership" "dev_bob" {
  team_id  = gitea_team.developers.id
  username = "bob"
}
```

**Import commands:**
```bash
# Import team (find team ID from Gitea UI or API)
terraform import gitea_team.developers 42

# Import repository assignments
terraform import gitea_team_repository.dev_backend acme/developers/backend
terraform import gitea_team_repository.dev_frontend acme/developers/frontend
terraform import gitea_team_repository.dev_docs acme/developers/docs

# Import memberships
terraform import gitea_team_membership.dev_alice 42/alice
terraform import gitea_team_membership.dev_bob 42/bob
```
