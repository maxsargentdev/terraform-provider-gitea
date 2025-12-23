# terraform-provider-gitea

⚠️ Vibe coded project !⚠️

A terraform provider for gitea generated from its openAPI spec an alternative to the go-gitea/terraform-provider-gitea provider.

First followed documentation on code generation: https://developer.hashicorp.com/terraform/plugin/code-generation

Then setup a basic local development loop with docker compose and some vscode tasks.

Then recruited claude to vibe code and test against terraform / giteas API.

## Generating schemas from openapi spec

* 1 - Download v2 oas yaml schema for gitea from `https://docs.gitea.com/api/<version>` e.g. `https://docs.gitea.com/api/1.25` and save in a versioned folder with the name `openapi_2.yaml`
* 2 - Convert that v2 oas schema into v3 oas schema
* 3 - Write or reuse `generator_config.yaml` in that directory to map the gitea API

The result of this is code we can use in the provider which makes for less work.

This is how I setup the initial schema:

~~~bash
docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate -i /local/openapi_2.yaml -g openapi -o /local/tmp

mv tmp/openapi.json ./openapi_3.json

rm -rf tmp

sed -i 's|#/definitions/|#/components/schemas/|g' openapi_3.json
~~~

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

`C:\Users\<USER>\go\bin`

~~~
provider_installation {
    dev_overrides {
        "hashicorp.com/maxsargentdev/gitea" = "C:\Users\SargentM\go\bin"
    }
    direct {}
}
~~~
