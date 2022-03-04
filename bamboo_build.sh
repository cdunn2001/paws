#!/usr/bin/bash

source env.sh

set -vex
which go

make test
make vet
make build
bin/paws -h
