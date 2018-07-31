#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/flokkr/flokkr-operator/pkg/generated \
github.com/flokkr/flokkr-operator/pkg/apis \
flokkr:v1alpha1 \
--go-header-file "./tmp/codegen/boilerplate.go.txt"
