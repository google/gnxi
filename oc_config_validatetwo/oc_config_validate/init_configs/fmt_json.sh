#!/bin/bash

# Format the JSON files

BASEDIR=$(dirname $0)

for f in "${BASEDIR}"/*.json; do
  if python3 -m json.tool --indent 2 "$f" "${f}.pprint" 2> /dev/null; then
    mv "${f}.pprint" "$f"
  else
    rm "${f}.pprint"
  fi 
done
