#!/usr/bin/env bash

helm init --client-only
NAMESPACE=$1
CRD=$2
helm delete --purge $CRD