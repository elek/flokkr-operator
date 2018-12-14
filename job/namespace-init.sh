!/usr/bin/env bash

helm init --client-only
helm repo add flokkr https://flokkr.github.io/charts
helm repo update
helm search flokkr
NAMESPACE=$1

echo "Initializing new flokkr namespace"
set -ex
helm upgrade --install --namespace $NAMESPACE $NAMESPACE-prometheus flokkr/prometheus
helm upgrade --install --namespace $NAMESPACE $NAMESPACE-grafana flokkr/grafana --set prometheusurl=http://$NAMESPACE-prometheus:9090
helm upgrade --install --namespace $NAMESPACE flokkr-$NAMESPACE flokkr/namespace
