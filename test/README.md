# Test Harness

This directory contains a complete integration test harness for the Gitea Terraform provider.

## Purpose

Unlike the simple examples in `examples/`, this directory contains a comprehensive test setup that:

- **Tests multiple resources together** - Validates complex scenarios with dependencies between resources
- **Uses Docker Compose** - Runs a local Gitea instance for testing
- **Verifies end-to-end workflows** - Tests the full lifecycle of resources (create, read, update, delete)
- **Validates data sources** - Ensures data sources correctly retrieve information

## Prerequisites

- Docker and Docker Compose
- Terraform CLI
- Go (for building the provider)

## Quick Start

### 1. Start Gitea and Apply Configuration

Use the VSCode task or run manually:

```powershell
# Using VSCode task
Run Task: "Install and Apply"

# Or manually
cd test
docker compose up -d
cd ..
go install
cd test
terraform init
terraform apply -auto-approve
```

### 2. Verify Resources

Check the outputs to see created resources:

```powershell
terraform output
```

### 3. Clean Up

```powershell
# Using VSCode task
Run Task: "Clean Gitea"

# Or manually
cd test
terraform destroy -auto-approve
docker compose down -v
Remove-Item terraform.tfstate*
```

## What's Tested

The test harness validates:

### Resources
- **gitea_user** - User creation with email and password
- **gitea_org** - Organization with full metadata
- **gitea_repository** - Repository creation (defaults to root user)
- **gitea_team** - Team with fine-grained permissions using `units_map`
- **gitea_team_membership** - Adding users to teams
- **gitea_branch_protection** - Branch protection rules with whitelists
- **gitea_token** - API token generation for users

### Data Sources
- **gitea_user** - Looking up users by username
- **gitea_org** - Looking up organizations
- **gitea_team** - Looking up teams by ID
- **gitea_team_membership** - Verifying team membership
- **gitea_repository** - Looking up repositories
- **gitea_branch_protection** - Looking up branch protection rules

### Complex Scenarios
- Cross-resource dependencies (e.g., team depends on org, membership depends on team and user)
- Resource outputs used as inputs to other resources
- Data sources reading from created resources

## Configuration

The test harness uses:

- **Gitea Instance**: `http://localhost:3000`
- **Admin User**: `root` / `admin1234`
- **Docker Compose**: Configured in `docker-compose.yaml`

## VSCode Tasks

Several tasks are available for common workflows:

- **Start Gitea** - Start the Docker Compose Gitea instance
- **Stop Gitea** - Stop the Gitea instance
- **Clean Gitea** - Remove volumes and state files (full cleanup)
- **Go Install** - Build and install the provider locally
- **Terraform Plan** - Preview changes
- **Terraform Apply** - Apply changes
- **Terraform Destroy** - Destroy all resources
- **Install and Apply** - Full workflow from start to finish
- **Terraform Cycle** - Plan → Apply → Destroy sequence

## Differences from `examples/`

| Aspect | `test/` | `examples/` |
|--------|---------|-------------|
| **Purpose** | Integration testing | Documentation |
| **Complexity** | Multi-resource scenarios | Single resource demos |
| **Dependencies** | Many cross-resource deps | Minimal/standalone |
| **Environment** | Local Docker Compose | Generic/production-like |
| **Maintenance** | Updated with new features | Simple, stable examples |

## Adding New Tests

When adding a new resource or data source:

1. Add a `<resource_name>.tf` file to this directory
2. Include resource creation and any related data sources
3. Add outputs to verify the resource works
4. Test dependencies with existing resources if relevant
5. Update this README if adding complex scenarios

## Troubleshooting

### Gitea not responding
Wait a few seconds after `docker compose up -d` for Gitea to fully start.

### Resource already exists errors
Run the Clean Gitea task to reset the environment.

### Provider not found
Run `go install` from the repository root.

### Terraform state issues
Remove `terraform.tfstate*` files and start fresh.
