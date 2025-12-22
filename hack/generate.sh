#!/bin/bash
REPO_ROOT=$(git rev-parse --show-toplevel)

pushd $REPO_ROOT/$1

docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate -i /local/openapi_2.yaml -g openapi -o /local/tmp

mv tmp/openapi.json ./openapi_3.json

rm -rf tmp

tfplugingen-openapi generate \
    --config generator_config.yaml \
    --output ./provider_code_spec.json openapi_3.json

tfplugingen-framework generate all \
    --input provider_code_spec.json \
    --output ./test

popd