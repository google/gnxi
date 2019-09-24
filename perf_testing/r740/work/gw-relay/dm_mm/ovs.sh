mkdir -p db
mkdir -p log
mkdir -p run

export OVS_DBDIR=$PWD/db
export OVS_RUNDIR=$PWD/run
export OVS_LOGDIR=$PWD/log

/usr/share/openvswitch/scripts/ovs-ctl --system-id=random start

ovs-vsctl add-br dm_mm
ovs-vsctl add-port dm_mm gre2 -- set Interface gre2 ofport_request=200
ovs-vsctl add-port dm_mm veth3 -- set Interface veth3 ofport_request=201
ip l set gre2 up
ip l set veth3 up
ovs-ofctl -OOpenFlow13 del-flows dm_mm
ovs-ofctl -OOpenFlow13 add-flow dm_mm priority=120,in_port=200,actions=output:201
ovs-ofctl -OOpenFlow13 add-flow dm_mm priority=120,in_port=201,actions=output:200
