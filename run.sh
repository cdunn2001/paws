#!/bin/bash
set -vex
P=23632
E=localhost:$P
mkdir -p tmp
rm -f tmp/*
F="-i -f"

curl $f -X GET $E/sockets
# Should print [1,2,3,4] quoted.


# darkcalFileUrl=file:/data/nrta/0/darkcal.h5
#curl $F -X POST -d '{"calibFileUrl": ""}' $E/sockets/1/darkcal/start

###
# Sleep after starting so they will finish before reset.

curl $F -X GET $E/sockets/1/darkcal
curl $F -X POST -d @sims/darkcal.start.json $E/sockets/1/darkcal/start
curl $F -X GET $E/sockets/1/darkcal
sleep 0.1
curl $F -X GET $E/sockets/1/darkcal
curl $F -X POST $E/sockets/1/darkcal/reset
curl $F -X GET $E/sockets/1/darkcal

curl $F -X GET $E/sockets/1/loadingcal
curl $F -X POST -d @sims/loadingcal.start.json $E/sockets/1/loadingcal/start
sleep 0.1

curl $F -X GET $E/sockets/1/basecaller
curl $F -X POST -d @sims/basecaller.start.json $E/sockets/1/basecaller/start
sleep 0.1
curl $F -X POST $E/sockets/1/basecaller/reset

curl $F -X GET $E/postprimaries
curl $F -X POST -d @sims/baz2bam.start.json $E/postprimaries
curl $F -X GET $E/postprimaries
sleep 0.1
curl $F -X GET $E/postprimaries
curl $F -X DELETE $E/postprimaries/m123

exit 0
###
# Stall the processes so we have time to stop them while RUNNING.

curl $F -X POST -d @sims/darkcal.start.json $E/sockets/2/darkcal/start?stall=5
curl $F -X GET $E/sockets/2/darkcal
curl $F -X POST $E/sockets/2/darkcal/stop
curl $F -X GET $E/sockets/2/darkcal
curl $F -X POST $E/sockets/2/darkcal/reset
curl $F -X GET $E/sockets/2/darkcal

curl $F -X POST -d @sims/loadingcal.start.json $E/sockets/2/loadingcal/start?stall=5
curl $F -X POST $E/sockets/2/loadingcal/stop
curl $F -X POST $E/sockets/2/loadingcal/reset

curl $F -X POST -d @sims/basecaller.start.json $E/sockets/2/basecaller/start?stall=5
curl $F -X POST $E/sockets/2/basecaller/stop
curl $F -X POST $E/sockets/2/basecaller/reset

curl $F -X GET $E/postprimaries
curl $F -X POST -d @sims/baz2bam.start.json $E/postprimaries?stall=5
curl $F -X GET $E/postprimaries
curl $F -X POST $E/postprimaries/m123/stop
curl $F -X DELETE $E/postprimaries/m123
