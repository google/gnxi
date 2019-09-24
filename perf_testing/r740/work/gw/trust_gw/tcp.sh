#!/bin/bash

set -e 

runs=10
addr=104.134.241.60

echo "gateway to relay"
for i in $(seq 1 $runs); do iperf3 -c $addr -J | jq -r .end.sum_received.bits_per_second; done | awk '{n += $1}; END{printf "%.2e\n", n/NR}'

echo "relay to gateway"
for i in $(seq 1 $runs); do iperf3 -c $addr -R -J | jq -r .end.sum_received.bits_per_second; done | awk '{n += $1}; END{printf "%.2e\n", n/NR}'