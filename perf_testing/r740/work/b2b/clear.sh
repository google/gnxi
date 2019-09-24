#!/bin/bash

ipsec stop

ovs-vsctl del-br br0
ip netns del host

ip l d gre1
ip l d veth1
ip a d 104.134.241.62/24 dev eno2
