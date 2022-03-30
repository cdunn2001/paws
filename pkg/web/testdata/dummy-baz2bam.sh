#!/bin/bash
# https://stackoverflow.com/a/14203146

POSITIONAL_ARGS=()
FD=2
LOG_OUTPUT="default.dummy-baz2bam.$$.log"
PREFIX_OUT="foo"

while [[ $# -gt 0 ]]; do
  case $1 in
    --version) # maybe?
      VERSION=1
      shift # past argument
      ;;
    -o)
      PREFIX_OUT="$2"
      shift # past argument
      shift # past value
      ;;
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

if [[ ! -z "${VERSION}" ]]
then
    echo "0.0.0"
    exit 0
fi

BAZFILE=${POSITIONAL_ARGS[0]}
#BAMFILE=${BAZFILE%.baz}.bam
BAMFILE=${PREFIX_OUT}.bam
echo "BAZFILE=${BAZFILE}"
echo "BAMFILE=${BAMFILE}"

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
INFO | PAWNEE_STATUS {"state": "progress", "stageNumber": $1, "stageName": "$2", "counter": $3, "timeoutForNextStatus": $4, "stageWeighting": "$STAGE_WEIGHTING", "timestamp": "$TIMESTAMP"}
EOF
}

function count {
    for i in $(seq 1 ${STATUS_COUNT}); do
        report_status 1 "baz2bam" $i $STATUS_TIMEOUT
        sleep $STATUS_DELAY_SECONDS
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
