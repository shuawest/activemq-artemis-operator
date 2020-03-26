#!/bin/sh
REGISTRY=localhost.localdomain:5000/activemq-artemis
IMAGE=activemq-artemis-operator
TAG=0.13.0-continuity
CFLAGS="--redhat --build-tech-preview"

go generate ./...
if [[ -z ${CI} ]]; then
    ./scripts/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
   
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o build/_output/bin/activemq-artemis-operator github.com/rh-messaging/activemq-artemis-operator/cmd/manager
fi



