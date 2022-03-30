#!/bin/bash

pwd
nproc
ls -larth ..
ls -larth
source ./env.sh
set -vex

make -C rpm build-rpm
find ./rpm/

ls -larth ./rpm/BUILD/pa-wsgo.rpm.tar
tar tvf ./rpm/BUILD/pa-wsgo.rpm.tar
