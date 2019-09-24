#!/bin/bash

ip netns add office
ip netns add dm_carl

ip link add veth1 type veth peer name veth2
ip link set veth1 netns office
ip link set veth2 netns dm_carl

ip netns exec office ip a a 10.10.10.10/24 dev veth1
ip netns exec office ip l set dev veth1 mtu 1370
ip netns exec office ip l set veth1 up
ip netns exec office ip l set lo up

ip link set enp6s0f1 netns dm_carl

ip netns exec dm_carl ip a a 104.134.241.60/29 dev enp6s0f1
ip netns exec dm_carl ifconfig enp6s0f1 up
ip netns exec dm_carl ip r add default via 104.134.241.62 dev enp6s0f1
ip netns exec dm_carl ip l add gre1 type gretap local 104.134.241.60 remote 104.134.6.15 dev enp6s0f1 ttl 32
cd dm_carl
ip netns exec dm_carl ./ovs.sh
cd -
