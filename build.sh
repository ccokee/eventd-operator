#!/bin/bash

# Build the project
echo "Building the project..."
go mod tidy && make generate && make manifests

echo "Building the docker image..."
make docker-build

echo "Pushing the docker image..."
make docker-push

