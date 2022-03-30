#!/bin/bash

pwd
ls -larth ..
ls -larth
source ./env.sh
set -vex

make -C rpm build-rpm
find ./rpm/

ls -larth ./rpm/BUILD/pa-wsgo.rpm
