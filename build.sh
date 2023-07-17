#!/bin/bash

# Build the project and push the docker image to the registry
# Usage: ./build.sh <image_tag_base> <version>

echo "Building the project..."
go mod tidy && make generate && make manifests

echo "Building and pushing the docker image..."
IMAGE_TAG_BASE=$1 VERSION=$2 make docker-build docker-push

