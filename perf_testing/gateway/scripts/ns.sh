#! /bin/bash

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as sudo or root."
  exit 1
fi

/home/esdn/scripts/node.sh add iperf eno2 eno3
