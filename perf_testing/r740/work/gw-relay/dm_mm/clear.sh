export OVS_DBDIR=$PWD/db
export OVS_RUNDIR=$PWD/run
export OVS_LOGDIR=$PWD/log

ipsec stop
ovs-vsctl del-br dm_mm
/usr/share/openvswitch/scripts/ovs-ctl stop
