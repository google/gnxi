#!/bin/bash

set -e

runs=100
addr=10.10.10.11

echo "gateway to relay"
for i in $(seq 1 $runs); do iperf3 -c $addr -u -b 10G -P 5 -J | jq -r '.end.sum.bits_per_second, .end.sum.lost_percent'; echo ""; done | awk 'BEGIN{ RS=""; FS = "\n" }{bw+=$1; loss+=$2}END{printf "%.2e, %.2f\n", bw/NR, loss/NR}'

echo "relay to gateway"
for i in $(seq 1 $runs); do iperf3 -c $addr -u -b 10G -P 5 -R -J | jq -r '.end.sum.bits_per_second, .end.sum.lost_percent'; echo ""; done | awk 'BEGIN{ RS=""; FS = "\n" }{bw+=$1; loss+=$2}END{printf "%.2e, %.2f\n", bw/NR, loss/NR}'

