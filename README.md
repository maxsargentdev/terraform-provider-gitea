# terraform-provider-openapi-gitea

A terraform provider for gitea generated from its openAPI spec.

## Generating resources from openapi spec

1 - Download v2 oas yaml schema for gitea from `https://docs.gitea.com/api/<version>` e.g. `https://docs.gitea.com/api/1.25` and save in a versioned folder with the name `openapi_2.yaml`
2 - Write or reuse `generator_config.yaml` in that directory to map the gitea API
3 - Generate: `./hack/generate.sh <version>` e.g. `./hack/generate.sh 1.25.3`

The result of this is code we can use in the provider.

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
