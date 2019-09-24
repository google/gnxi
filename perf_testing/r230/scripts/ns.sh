#! /bin/bash

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as sudo or root."
  exit 1
fi

/home/esdn/scripts/node.sh add iperf enp6s0f0 enp6s0f1
