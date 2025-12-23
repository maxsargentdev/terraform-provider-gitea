#!/bin/bash
REPO_ROOT=$(git rev-parse --show-toplevel)

pushd $REPO_ROOT/openapi/$1

# docker run --rm -v "$PWD:/local" openapitools/openapi-generator-cli generate -i /local/openapi_2.yaml -g openapi -o /local/tmp

# mv tmp/openapi.json ./openapi_3.json

# rm -rf tmp

# sed -i 's|#/definitions/|#/components/schemas/|g' openapi_3.json

tfplugingen-openapi generate \
    --config generator_config.yaml \
    --output ./provider_code_spec.json openapi_3.json

tfplugingen-framework generate resources \
    --input provider_code_spec.json \
    --output $REPO_ROOT/internal

tfplugingen-framework generate data-sources \
    --input provider_code_spec.json \
    --output $REPO_ROOT/internal

popd