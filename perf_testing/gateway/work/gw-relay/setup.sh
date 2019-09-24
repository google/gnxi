ip netns add office
ip netns add dm_carl
ip netns add trust_gw
ip netns add dm_mm
ip netns add nfvs_sw

ip link add veth1 type veth peer name veth2
ip link add veth3 type veth peer name veth4
ip link add todemux type veth peer name ipsecp

ip link set veth1 netns office
ip link set veth2 netns dm_carl
ip link set todemux netns trust_gw
ip link set ipsecp netns dm_mm
ip link set veth3 netns dm_mm
ip link set veth4 netns nfvs_sw
ip link set eno2 netns dm_carl
ip link set eno3 netns trust_gw

ip netns exec office ip a a 10.10.10.10/24 dev veth1
ip netns exec office ip link set dev veth1 mtu 1300
ip netns exec office ip link set veth1 up

ip netns exec nfvs_sw ip a a 10.10.10.11/24 dev veth4
ip netns exec nfvs_sw ip link set dev veth4 mtu 1300
ip netns exec nfvs_sw ip link set veth4 up

ip netns exec dm_carl ip a a 104.134.241.60/29 dev eno2
ip netns exec dm_carl ifconfig eno2 up
ip netns exec dm_carl ip r a default via 104.134.241.62 dev eno2
ip netns exec dm_carl ip link add gre1 type gretap local 104.134.241.60 remote 104.134.6.15 dev eno2 ttl 32
cd dm_carl
ip netns exec dm_carl ./ovs.sh
cd -

ip netns exec trust_gw ip a a 104.134.241.62/29 dev eno3
ip netns exec trust_gw ip a a 169.254.0.117/30 dev todemux
ip netns exec trust_gw ifconfig eno3 up
ip netns exec trust_gw ip l set todemux up
ip netns exec trust_gw ip r a 104.134.6.15 via 169.254.0.118 dev todemux
ip netns exec trust_gw sysctl -w net.ipv4.ip_forward=1

ip netns exec dm_mm ip a a 169.254.0.118/30 dev ipsecp
ip netns exec dm_mm ip l set ipsecp up
ip netns exec dm_mm ip r a 104.134.241.60 via 169.254.0.117 dev ipsecp
ip netns exec dm_mm ip a a 104.134.6.15/32 dev lo
ip netns exec dm_mm ip l set lo up
ip netns exec dm_mm ip l a gre2 type gretap local 104.134.6.15 remote 104.134.241.60 dev ipsecp ttl 32
cd dm_mm
ip netns exec dm_mm ./ovs.sh
cd -