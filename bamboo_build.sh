#!/bin/bash

set -vex
pwd
ls -larth ..
ls -larth

echo "SHELL=$SHELL"
echo "BASH_VERSION=$BASH_VERSION"
source ./env.sh

which go

make test
make vet
make build
bin/pawsgo -h
