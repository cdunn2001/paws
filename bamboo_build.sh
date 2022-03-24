#!/bin/bash

pwd
ls -larth ..
ls -larth
source ./env.sh
set -vex

echo "SHELL=$SHELL"
echo "BASH_VERSION=$BASH_VERSION"

which go

go clean -testcache ./pkg/web/...
make test
make vet
make build
bin/pawsgo -h
