# terraform-provider-openapi-gitea

A terraform provider for gitea generated from its openAPI spec.

## Steps

1 - Download v2 yaml schema for gitea from `https://docs.gitea.com/api/<version>` e.g. `https://docs.gitea.com/api/1.25` and save in a versioned folder with the name `openapi_2.yaml`
2 - Write a `generator_config.yaml` in that directory to map the gitea API
2 - Generate: `./hack/generate.sh <version>` e.g. `./hack/generate.sh 1.25.3`