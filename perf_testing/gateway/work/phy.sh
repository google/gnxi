#!/bin/bash

ip netns add ns1
ip netns add ns2

ip link set eno2 netns ns1
ip link set eno3 netns ns2

ip netns exec ns1 ip addr add 1.1.1.1/24 dev eno2
ip netns exec ns1 ip link set eno2 up
ip netns exec ns2 ip addr add 1.1.1.2/24 dev eno3
ip netns exec ns2 ip link set eno3 up