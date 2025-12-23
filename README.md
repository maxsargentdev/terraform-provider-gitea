# terraform-provider-gitea

A terraform provider for gitea generated from its openAPI spec an alternative to the go-gitea/terraform-provider-gitea provider.

This is also 

## Generating schemas from openapi spec

1 - Download v2 oas yaml schema for gitea from `https://docs.gitea.com/api/<version>` e.g. `https://docs.gitea.com/api/1.25` and save in a versioned folder with the name `openapi_2.yaml`
2 - Conver that v2 oas schema into v3 oas schema
2 - Write or reuse `generator_config.yaml` in that directory to map the gitea API
3 - Generate: `./hack/generate.sh <version>` e.g. `./hack/generate.sh 1.25.3` to output schemas and data model types for Gitea

The result of this is code we can use in the provider which makes for less work.

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

Currently I have worked around issues with the existing provider and the setting of team permissions, as well as missing resources for org level secrets.

## .terraformrc

C:\Users\<USER>\go\bin

~~~
provider_installation {
    dev_overrides {
        "hashicorp.com/maxsargentdev/gitea" = "C:\Users\SargentM\go\bin"
    }
    direct {}
}
~~~
