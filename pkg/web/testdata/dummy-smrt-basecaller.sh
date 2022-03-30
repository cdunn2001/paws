#!/bin/bash
# https://stackoverflow.com/a/14203146

POSITIONAL_ARGS=()
FD=2
LOG_OUTPUT="default.dummy-smrt-basecaller.$$.log"
OUTPUT_TRC_FILE=""
OUTPUT_BAZ_FILE=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --statusfd)
      FD="$2"
      shift # past argument
      shift # past value
      ;;
    --logfilter)
      LOG_FILTER="$2"
      shift # past argument
      shift # past value
      ;;
    --logoutput)
      LOG_OUTPUT="$2"
      shift # past argument
      shift # past value
      ;;
    --outputbazfile)
      OUTPUT_BAZ_FILE="$2"
      shift # past argument
      shift # past value
      ;;
    --outputtrcfile)
      OUTPUT_TRC_FILE="$2"
      shift # past argument
      shift # past value
      ;;
    --config)
      CONFIG="$2"
      shift # past argument
      shift # past value
      ;;
    --maxFrames)
      MAX_FRAMES="$2"
      shift # past argument
      shift # past value
      ;;
    --numWorkerThreads)
      NUM_WORKER_THREADS="$2"
      shift # past argument
      shift # past value
      ;;
    --version) # maybe?
      VERSION=1
      shift # past argument
      ;;
    --*)
      echo "Unknown option $1 $2"
      shift # past argument
      shift # past value
      #exit 1 # Fine.
      ;;
    -*)
      echo "Unknown flag $1"
      shift # past argument
      #exit 1 # Fine.
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

if [[ ! -z "${VERSION}" ]]
then
    echo "0.0.0"
    exit 0
fi

# Optional env-vars:
: "${STATUS_COUNT:=0}"
: "${STATUS_DELAY_SECONDS:=0.0}"
DOUBLE_DELAY=$(perl -e "print $STATUS_DELAY_SECONDS * 2.0")
: "${STATUS_TIMEOUT:=$DOUBLE_DELAY}"

TIMESTAMP="20220223T146198.099Z" # arbitrary
STAGE_WEIGHTING="[0, 100, 0]"

function log {
    echo "$1" >> "${LOG_OUTPUT}"
}

function report_status {
    # ARGS: number, name, counter, next
    # Not reported: counterMax
    # Do we need "timestamp"?
    cat >&$FD << EOF
INFO | SMRT_BASECALLER_STATUS {"state": "progress", "stageNumber": $1, "stageName": "$2", "counter": $3, "timeoutForNextStatus": $4, "stageWeighting": "$STAGE_WEIGHTING", "timestamp": "$TIMESTAMP", "ready": $5}
EOF
}

function count {
    for i in $(seq 1 ${STATUS_COUNT}); do
        report_status 1 "baz2bam" $i $STATUS_TIMEOUT true
        sleep $STATUS_DELAY_SECONDS
    done
}

set -vex

report_status 0 "init" 0 2 false
sleep $STATUS_DELAY_SECONDS
count
report_status 2 "fini" 0 2 false
sleep $STATUS_DELAY_SECONDS

touch ${LOG_OUTPUT}
if [[ $OUTPUT_BAZ_FILE != "" ]]
then
  touch ${OUTPUT_BAZ_FILE}
fi  
if [[ $OUTPUT_TRC_FILE != "" ]]
then
  touch ${OUTPUT_TRC_FILE}
fi
