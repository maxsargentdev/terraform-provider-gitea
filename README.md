# terraform-provider-gitea

⚠️ Vibe coded project !⚠️

A terraform provider for gitea generated from its openAPI spec an alternative to the go-gitea/terraform-provider-gitea provider.

First followed documentation on code generation: https://developer.hashicorp.com/terraform/plugin/code-generation.

Then set up a basic local development loop with docker compose and some vscode tasks.

Then recruited claude to vibe code and test against terraform / the Gitea API.

## Generating schemas from openapi spec

* 1 - Download v2 oas yaml schema for gitea from `https://docs.gitea.com/api/<version>` e.g. `https://docs.gitea.com/api/1.25` and save in a versioned folder with the name `openapi_2.yaml`
* 2 - Convert that v2 oas schema into v3 oas schema
* 3 - Write or reuse `generator_config.yaml` in that directory to map the gitea API

The result of this is code we can use in the provider which makes for less work.

This is how I set up the initial schema:

~~~bash
docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate -i /local/openapi_2.yaml -g openapi -o /local/tmp

mv tmp/openapi.json ./openapi_3.json

rm -rf tmp

sed -i 's|#/definitions/|#/components/schemas/|g' openapi_3.json
~~~

## Vendoring

We had to `go mod vendor` the gitea SDK and make some changes to support units_map to allow the correct setting of permissions.

The modifications are committed directly to the vendored code in `vendor/code.gitea.io/sdk/gitea/org_team.go`.

Patch files to any vendored code are in vendor-patch in case claude rewrites anything.

## Priorities

I am aiming to use this to replace my own usage of that provider so will start by recreating the following resources:

- gitea_token
- gitea_org
- gitea_repository
- gitea_repository_branch_protection
- gitea_user
- gitea_team
- gitea_team_members
- gitea_team_membership

Aiming to resolve issues with gitea_team permissions that exist in the current provider.

Currently, I have worked around issues with the existing provider and the setting of team permissions, as well as missing resources for org level secrets.

## Future Resources

### High Priority

#### Repository Management
- **gitea_deploy_key** - Deploy keys for repositories (read-only SSH keys)
- **gitea_repository_collaborator** - Manage repository collaborators with permissions
- **gitea_repository_webhook** - Repository webhooks (Discord, Slack, etc.)
- **gitea_repository_label** - Custom labels for issues/PRs
- **gitea_repository_milestone** - Project milestones
- **gitea_repository_release** - Manage releases and tags
- **gitea_tag_protection** - Protect specific tags from deletion/modification

#### User Management
- **gitea_user_key** - SSH public keys for users
- **gitea_user_gpg_key** - GPG keys for commit signing
- **gitea_user_email** - Additional email addresses for users

#### Organization Management
- **gitea_org_webhook** - Organization-level webhooks
- **gitea_org_label** - Organization-wide labels
- **gitea_org_secret** - Organization secrets for actions

#### Actions/CI
- **gitea_repository_secret** - Repository secrets for Gitea Actions
- **gitea_repository_variable** - Repository variables for Actions
- **gitea_org_variable** - Organization variables for Actions

#### Team Management
- **gitea_team_repository** - Assign repositories to teams with permissions

### Medium Priority

#### Repository Features
- **gitea_repository_mirror** - Mirror external repositories
- **gitea_repository_topic** - Repository topics/tags
- **gitea_repository_transfer** - Transfer repository ownership

#### Advanced Features
- **gitea_oauth_application** - OAuth2 applications
- **gitea_cron_task** - Admin cron tasks (data source/trigger)

#### Packages
- **gitea_package** - Package information (data source only - packages are published via package managers)

## .terraformrc

`C:\Users\<USER>\go\bin`

~~~
provider_installation {
    dev_overrides {
        "hashicorp.com/maxsargentdev/gitea" = "C:\Users\SargentM\go\bin"
    }
    direct {}
}
~~~
