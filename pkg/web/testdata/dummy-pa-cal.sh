#!/bin/bash
# https://stackoverflow.com/a/14203146

POSITIONAL_ARGS=()
FD=2

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

# Optional env-vars:
: "${STATUS_COUNT:=0}"
: "${STATUS_DELAY_SECONDS:=0.0}"

TIMESTAMP="20220223T146198.099Z" # arbitrary
STAGE_WEIGHTING="[0, 100, 0]"

function report_status {
    # ARGS: number, name, counter, next
    # Not reported: counterMax
    # Do we need "timestamp"?
    cat >&$FD << EOF
INFO | PA_CAL_STATUS {"state": "progress", "stageNumber": $1, "stageName": "$2", "counter": $3, "timeToNextStatus": $4, "stageWeighting": "$STAGE_WEIGHTING", "timestamp": "$TIMESTAMP"}
EOF
}

function count {
    for i in $(seq 1 ${STATUS_COUNT}); do
        sleep $STATUS_DELAY_SECONDS
        report_status 1 "pa-cal" $i $STATUS_DELAY_SECONDS
    done
}

set -vex

report_status 0 "init" 0 1
count
report_status 2 "fini" 0 1

touch ${LOG_OUTPUT}
touch ${OUTPUT_FILE}
