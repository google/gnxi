#! /bin/bash

cd dm_mm
ip netns exec dm_mm ./clear.sh
cd -

ip netns exec trust_gw ip link set eno3 netns 1

ip netns del trust_gw
ip netns del dm_mm
ip netns del nfvs_sw
