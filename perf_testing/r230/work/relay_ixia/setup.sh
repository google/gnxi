#!/bin/bash

ip netns add dm_carl

ip link set enp6s0f1 netns dm_carl
ip link set enp6s0f0 netns dm_carl

ip netns exec dm_carl ip a a 104.134.241.60/29 dev enp6s0f1
ip netns exec dm_carl ifconfig enp6s0f1 up
ip netns exec dm_carl ifconfig enp6s0f0 up
ip netns exec dm_carl ip r add default via 104.134.241.62 dev enp6s0f1
ip netns exec dm_carl ip l add gre1 type gretap local 104.134.241.60 remote 104.134.6.15 dev enp6s0f1 ttl 32
cd dm_carl
ip netns exec dm_carl ./ovs.sh
cd -
