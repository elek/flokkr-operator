#!/usr/bin/env bash

###
# Create the docker image
###
build(){
   docker build -t flokkr/flokkr-operator .
}

###
# Deploy the docker image to the docker hub
###
deploy(){
   docker push flokkr/flokkr-operator
}

$@

