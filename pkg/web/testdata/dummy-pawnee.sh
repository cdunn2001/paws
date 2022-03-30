#!/bin/bash
# https://stackoverflow.com/a/14203146

POSITIONAL_ARGS=()
FD=2
LOG_OUTPUT="default.pawnee.$$.log"

while [[ $# -gt 0 ]]; do
  case $1 in
    --version)
      VERSION=1
      shift # past argument
      ;;
    --status-fd)
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
    --rmd)
      RMD="$2"
      shift # past argument
      shift # past value
      ;;
    --mess)
      MESS="$2"
      shift # past argument
      shift # past value
      ;;
    --ppaConfig)
      PPA_CONFIG="$2"
      shift # past argument
      shift # past value
      ;;
    --out-bash)
      OUT_BASH="$2"
      shift # past argument
      shift # past value
      ;;
    --run)
      RUN=1
      shift # past argument
      ;;
    --nohup)
      NOHUP=1
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
touch ${OUT_BASH}

# Technically, the xml is an output, but that is named
# in a config file. We can simply touch a hard-coded file that
# is expected by the tests.
#touch prefix.consensusreadset.xml
