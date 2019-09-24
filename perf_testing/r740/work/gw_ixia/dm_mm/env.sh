#!/bin/bash

export OVS_DBDIR=$PWD/db
export OVS_RUNDIR=$PWD/run
export OVS_LOGDIR=$PWD/log

mount -t overlay -o lowerdir=/etc,upperdir=etc,workdir=etc_workdir none /etc
mount -t overlay -o lowerdir=/var/run,upperdir=var_run,workdir=var_run_workdir none /var/run