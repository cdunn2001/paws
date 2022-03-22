#!/bin/bash
# https://stackoverflow.com/a/14203146
set -vex
pwd

POSITIONAL_ARGS=()
FD=2
LOG_OUTPUT="default.dummy-pa-cal.$$.log"
OUTPUT_FILE="default.output"

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
      --outputFile)
          OUTPUT_FILE="$2"
          shift # past argument
          shift # past value
          ;;
      --config)
          CONFIG="$2"
          shift # past argument
          shift # past value
          ;;
      --version)
          VERSION=1
          shift # past argument
          ;;
      --sra)
          shift # past argument
          shift # past value
          ;;
      --movieNum)
          shift # past argument
          shift # past value
          ;;
      --numFrames)
          shift # past argument
          shift # past value
          ;;
      --timeoutSeconds)
          shift # past argument
          shift # past value
          ;;
      --cal)
          CAL="$2"
          shift # past argument
          shift # past value
          ;;
      --inputDarkCalFile)
          shift # past argument
          inputDarkCalFile=$1
          shift # past value
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

function log {
    echo "$1" >> "${LOG_OUTPUT}"
}

function report_status {
    # ARGS: number, name, counter, next
    # Not reported: counterMax
    # Do we need "timestamp"?
    cat >&$FD << EOF
INFO | PA_CAL_STATUS {"state": "progress", "stageNumber": $1, "stageName": "$2", "counter": $3, "timeoutForNextStatus": $4, "stageWeighting": "$STAGE_WEIGHTING", "timestamp": "$TIMESTAMP"}
EOF
}

function count {
    for i in $(seq 1 ${STATUS_COUNT}); do
        sleep $STATUS_DELAY_SECONDS
        report_status 1 "pa-cal" $i $STATUS_DELAY_SECONDS
    done
}

function report_exception {
    echo "Exception: $1" >&2
    # ARGS: message
    cat >&$FD << EOF
ERROR | PA_CAL_STATUS {"state": "exception", "message": "$1"}
EOF
}

if [[ -z "${CAL}" || (${CAL} != "Dark" && ${CAL} != "Loading") ]]
then
    report_exception "--cal='${CAL}' must be Dark or Loading"
    exit 1
fi

if [[ ${CAL} == "Loading" && ! -f ${inputDarkCalFile} ]]
then
    report_exception "inputDarkCalFile ${inputDarkCalFile} does not exist (for --cal=Loading)"
    exit 1
fi

# Optional env-vars:
: "${STATUS_COUNT:=0}"
: "${STATUS_DELAY_SECONDS:=0.0}"

echo "STATUS_COUNT=$STATUS_COUNT"
echo "STATUS_DELAY_SECONDS=$STATUS_DELAY_SECONDS"
sleep "$STATUS_DELAY_SECONDS"

TIMESTAMP="20220223T146198.099Z" # arbitrary
STAGE_WEIGHTING="[0, 100, 0]"

#set -vex

date > ${LOG_OUTPUT}
log "Starting pa-cal"
report_status 0 "init" 0 1
count
report_status 2 "fini" 0 1

log "Ending pa-cal"
date >> ${LOG_OUTPUT}
touch ${OUTPUT_FILE}
