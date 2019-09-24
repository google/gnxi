#!/bin/bash

ip netns add host
ip l a veth1 type veth peer name veth2
ip l set veth2 netns host

ip a a 104.134.241.60/24 dev enp6s0f0
ifconfig enp6s0f0 up
ip l a gre1 type gretap local 104.134.241.60 remote 104.134.241.62 dev enp6s0f0

ovs-vsctl add-br br0
ovs-vsctl add-port br0 gre1 -- set Interface gre1 ofport_request=200
ovs-vsctl add-port br0 veth1 -- set Interface veth1 ofport_request=201
ip link set gre1 up mtu 1500
ip l set veth1 up
ovs-ofctl -OOpenFlow13 del-flows br0
ovs-ofctl -OOpenFlow13 add-flow br0 priority=120,in_port=200,actions=output:201
ovs-ofctl -OOpenFlow13 add-flow br0 priority=120,in_port=201,actions=output:200

ip netns exec host ip a a 10.10.10.11/24 dev veth2
ip netns exec host ip l set veth2 mtu 1370
ip netns exec host ip l set veth2 up

ipsec start
