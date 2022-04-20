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
go test ./pkg/... -v -timeout 30s -count=1
make vet
make release
bin/pawsgo --version
