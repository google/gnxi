#!/bin/bash

cd dm_carl
ip netns exec dm_carl ./clear.sh
cd -

ip netns exec dm_carl ip link set enp6s0f1 netns 1
ip netns exec dm_carl ip link set enp6s0f0 netns 1

ip netns del dm_carl
