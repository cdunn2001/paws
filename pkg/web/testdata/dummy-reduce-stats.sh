#!/bin/bash
# https://stackoverflow.com/a/14203146

POSITIONAL_ARGS=()
FD=2
LOG_OUTPUT="default.dummy-reduce-stats.$$.log"

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
    --metadata)
      METADATA="$2"
      shift # past argument
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

BAZFILE=${POSITIONAL_ARGS[0]}
BAMFILE=${BAZFILE%.baz}.bam
echo "BAZFILE=${BAZFILE}"
echo "BAMFILE=${BAMFILE}"

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
INFO | PAWNEE_STATUS {"state": "progress", "stageNumber": $1, "stageName": "$2", "counter": $3, "timeToNextStatus": $4, "stageWeighting": "$STAGE_WEIGHTING", "timestamp": "$TIMESTAMP"}
EOF
}

function count {
    for i in $(seq 1 ${STATUS_COUNT}); do
        sleep $STATUS_DELAY_SECONDS
        report_status 1 "baz2bam" $i $STATUS_DELAY_SECONDS
    done
}

set -vex

report_status 0 "init" 0 1
count
report_status 2 "fini" 0 1

touch ${LOG_OUTPUT}
touch ${BAMFILE}

# Technically, the xml is an output, but that is named
# in a config file. We can simply touch a hard-coded file that
# is expected by the tests.
#touch prefix.consensusreadset.xml
