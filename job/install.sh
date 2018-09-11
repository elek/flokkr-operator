#!/usr/bin/env bash

helm init --client-only
helm repo add flokkr https://flokkr.github.io/charts
helm repo update
helm search flokkr
NAMESPACE=$1
CRD=$2
kubectl get component -n $NAMESPACE -o json $CRD > component.json
cat component.json | jq -r '.spec.values' > values.json
python -c 'import sys, yaml, json; yaml.safe_dump(json.load(sys.stdin), sys.stdout, default_flow_style=False)' < values.json > values.yaml
TYPE=$(cat component.json | jq -r '.spec.type')
NAME=$(cat component.json | jq -r '.metadata.name')
echo "Executing helm install flokkr/$TYPE with values:"
cat values.yaml
helm upgrade --install $NAME flokkr/$TYPE -f values.yaml
