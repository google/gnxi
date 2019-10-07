#!/bin/bash

mount -t overlay -o lowerdir=/etc,upperdir=etc,workdir=etc_workdir none /etc
mount -t overlay -o lowerdir=/var/run,upperdir=var_run,workdir=var_run_workdir none /var/run
ipsec stop

export OVS_DBDIR=$PWD/db
export OVS_RUNDIR=$PWD/run
export OVS_LOGDIR=$PWD/log

ovs-vsctl del-br dm_mm
/usr/share/openvswitch/scripts/ovs-ctl stop
