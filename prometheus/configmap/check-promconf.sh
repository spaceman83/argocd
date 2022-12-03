#!/bin/bash
# $1 arg = File to be checked
# $2 arg = File to store output in
# $3 arg = url to post

ARGFILE=$(md5sum "$1")
COMPAREFILE=$(cat "$2")

if [ "$ARGFILE" = "$COMPAREFILE" ]
then
  sleep 1
else
  echo "File diff found, triggering prometheus reload"
  md5sum "$1" > "$2"
  curl --insecure -X POST "$3" -o /dev/null
fi