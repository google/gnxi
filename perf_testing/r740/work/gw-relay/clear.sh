cd dm_carl
ip netns exec dm_carl ./clear.sh
cd -
cd dm_mm
ip netns exec dm_mm ./clear.sh
cd -

ip netns exec trust_gw ip link set eno3 netns 1

ip netns del office
ip netns del trust_gw
ip netns del dm_carl
ip netns del dm_mm
ip netns del nfvs_sw
