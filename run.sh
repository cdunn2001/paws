#!/bin/bash
set -vex
P=23632
E=localhost:$P
mkdir -p tmp
rm -f tmp/*

curl -i -f -X GET $E/sockets
# Should print [1,2,3,4] quoted.

curl -i -f -X GET $E/sockets/1/darkcal

# darkcalFileUrl=file:/data/nrta/0/darkcal.h5
curl -i -f -X POST -d @sims/darkcal.start.json $E/sockets/1/darkcal/start
#curl -i -X POST -d '{"calibFileUrl": ""}' $E/sockets/1/darkcal/start

exit 0
# For later:
curl -i -f -X POST $E/sockets/1/darkcal/stop
curl -i -f -X POST $E/sockets/1/darkcal/reset

curl -i -f -X POST -d @sims/loadingcal.start.json $E/sockets/1/loadingcal/start
curl -i -f -X POST $E/sockets/1/loadingcal/stop
curl -i -f -X POST $E/sockets/1/loadingcal/reset

curl -i -f -X POST -d @sims/basecaller.start.json $E/sockets/1/basecaller/start
curl -i -f -X POST $E/sockets/1/basecaller/stop
curl -i -f -X POST $E/sockets/1/basecaller/reset

curl -i -f -X GET $E/postprimaries
curl -i -f -X POST -d @sims/baz2bam.start.json $E/postprimaries
curl -i -f -X GET $E/postprimaries
curl -i -f -X POST $E/postprimaries/m123/stop
curl -i -f -X DELETE $E/postprimaries/m123
