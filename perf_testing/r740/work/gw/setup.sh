#! /bin/bash

ip netns add trust_gw
ip netns add dm_mm
ip netns add nfvs_sw

ip link add dpu type veth peer name mm_dpu
ip link add todemux type veth peer name ipsecp

ip link set todemux netns trust_gw
ip link set ipsecp netns dm_mm
ip link set dpu netns dm_mm
ip link set mm_dpu netns nfvs_sw
ip link set eno3 netns trust_gw

ip netns exec nfvs_sw ip a a 10.10.10.11/24 dev mm_dpu
ip netns exec nfvs_sw ip l set dev mm_dpu mtu 1370
ip netns exec nfvs_sw ip link set mm_dpu up
ip netns exec nfvs_sw ip link set lo up

ip netns exec trust_gw ip addr add 104.134.241.62/29 dev eno3
ip netns exec trust_gw ip addr add 169.254.0.117/30 dev todemux
ip netns exec trust_gw ifconfig eno3 up
ip netns exec trust_gw ip link set todemux up
ip netns exec trust_gw ip route add 104.134.6.15 via 169.254.0.118 dev todemux
ip netns exec trust_gw sysctl -w net.ipv4.ip_forward=1

ip netns exec dm_mm ip addr add 169.254.0.118/30 dev ipsecp
ip netns exec dm_mm ip link set ipsecp up
ip netns exec dm_mm ip route add 104.134.241.60 via 169.254.0.117 dev ipsecp
ip netns exec dm_mm ip addr add 104.134.6.15/32 dev lo
ip netns exec dm_mm ip link set lo up
ip netns exec dm_mm ip link a gre2 type gretap local 104.134.6.15 remote 104.134.241.60 dev ipsecp ttl 32
cd dm_mm
ip netns exec dm_mm ./ovs.sh
cd -
