#!/usr/bin/env bash
helm init --client-only
helm repo add flokkr https://flokkr.github.io/charts
helm update
/go/bin/flokkr-operator
