#!/bin/bash
set -vex
P=23633
E=localhost:$P

curl -f -X GET $E/sockets
# Should print [1,2,3,4] quoted.

# darkcalFileUrl=file:/data/nrta/0/darkcal.h5
curl -f -X POST -d @sims/darkcal.start.json $E/sockets/1/darkcal/start
#curl -X POST -d '{"calibFileUrl": ""}' $E/sockets/1/darkcal/start
curl -f -X POST $E/sockets/1/darkcal/stop
curl -f -X POST $E/sockets/1/darkcal/reset

curl -f -X POST -d @sims/loadingcal.start.json $E/sockets/1/loadingcal/start
curl -f -X POST $E/sockets/1/loadingcal/stop
curl -f -X POST $E/sockets/1/loadingcal/reset

curl -f -X POST -d @sims/basecaller.start.json $E/sockets/1/basecaller/start
curl -f -X POST $E/sockets/1/basecaller/stop
curl -f -X POST $E/sockets/1/basecaller/reset
