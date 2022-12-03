#!/bin/bash
# $1 arg = File to be checked
# $2 arg = url to post

ARGFILE=$(md5sum "$1")
COMPAREFILE=$(cat "$1.md5")

echo $ARGFILE
echo $COMPAREFILE

if [ "$ARGFILE" = "$COMPAREFILE" ]
then
  sleep 1
else
  echo "File diff found, triggering prometheus reload"
  md5sum "$1" > "$1.md5"
  curl --insecure -X POST "$2"
fi