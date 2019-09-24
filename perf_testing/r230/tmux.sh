#!/bin/bash

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

  $TMUX new-window -t $SESSIONNAME:3 -n office
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX split-window -h -p 50 -t $SESSIONNAME:3
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX split-window -v -t $SESSIONNAME:3
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX send-keys 'htop' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:4 -n dm_carl
  $TMUX send-keys 'cd /home/esdn/work/dm_carl' 'C-m'
  $TMUX select-pane -t 0

  $TMUX new-window -t $SESSIONNAME:5 -n monitoring
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX split-window -h -t $SESSIONNAME:5
  $TMUX send-keys 'cd /home/esdn/work' 'C-m'
  $TMUX select-pane -t 0

  $TMUX select-window -t $SESSIONNAME:1
fi

$TMUX attach -t $SESSIONNAME
