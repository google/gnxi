#!/bin/sh

SESSIONNAME="work"
TMUX="tmux"

$TMUX has-session -t $SESSIONNAME > /dev/null 2>&1

if [ $? != 0 ]
then
  $TMUX new-session -d -s $SESSIONNAME -n src
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:2 -n test
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:3 -n trust_gw
  $TMUX send-keys 'cd /home/esdn/work/gw/trust_gw' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:4 -n dm_mm
  $TMUX send-keys 'cd /home/esdn/work/gw/dm_mm' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:5 -n nfvs_sw
  $TMUX send-keys 'cd /home/esdn/work/gw/nfvs_sw' 'C-m'
  $TMUX split-window -h -p 50 -t $SESSIONNAME:5
  $TMUX send-keys 'cd /home/esdn/work/gw/nfvs_sw' 'C-m'
  $TMUX split-window -v -t $SESSIONNAME:5
  $TMUX send-keys 'cd /home/esdn/work/gw/nfvs_sw' 'C-m'
  $TMUX send-keys 'htop' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:6 -n ns1
  $TMUX send-keys 'cd /home/esdn/work/ns1' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:7 -n ns2
  $TMUX send-keys 'cd /home/esdn/work/ns2' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:8 -n monitoring
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX split-window -v -t $SESSIONNAME:8
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX send-keys 'htop' 'C-m'
  $TMUX select-pane -t 0

  $TMUX select-window -t $SESSIONNAME:1
fi

$TMUX attach -t $SESSIONNAME