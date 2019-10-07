#!/bin/bash

# Usage of this script in machine running Rodete.
cmd_usage()
{
  echo "Usage Example:"
  echo "node add <node name> <interface> <interface> ..."
  echo "node delete <node name>"
  echo "node list"
}

# Checks if a namespace is present.
ns_present()
{
  ip netns | awk {'print $1'} | egrep "${1}"$ > /dev/null 2>&1
  return
}

# Adds a namespace and moves the interface(s).
add_ns()
{
  if [[ "$1" == "" ]] || [[ "$2" == "" ]]; then
    echo "Please provide node name and interface name."
    return 1
  fi
  # Create the namespace only if it is not present.
  ns_present "$1"
  if [[ $? != 0 ]]; then
    ip netns add "$1"
  fi
  # Add the interface if it is not already in the namespace.
  for LINK in "${@:2}"
  do
    ip netns exec "$1" ifconfig "$LINK" > /dev/null 2>&1
    if [[ "$?" != 0 ]]; then
      ip link set "$LINK" netns "$1"
      ip netns exec "$1" ifconfig "$LINK" up
    fi
  done
  return
}

# Deletes a namespace and moves the interface(s).
del_ns()
{
  if [[ "$1" == "" ]]; then
    echo "Please provide node name."
    return 1
  fi
  ns_present "$1"
  if [[ $? != 0 ]]; then
    echo "Namespace $1 does not exist"
    return 0
  fi
  LINKS=$(ip netns exec "$1" ifconfig -a | grep mtu | awk {'print $1'} | egrep -v 'lo' | sed 's/\:$//')
  for LINK in $LINKS
  do
    ip netns exec "$1" ifconfig "$LINK" down
    ip netns exec "$1" ip link set "$LINK" netns 1
  done
  ip netns delete "$1"
  return
}

# Lists all the namespaces.
list_ns()
{
  OUTPUT=$(ip netns | sort)
  if [[ "$OUTPUT" == "" ]]; then
    echo "No name spaces found"
  else
    echo "Name Spaces:"
    echo "$OUTPUT"
  fi
  return
}

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as sudo or root."
  exit 1
fi

case $1 in
  add)
    add_ns "${@:2}"
    exit $?
    ;;
  delete)
    del_ns "${@:2}"
    exit $?
    ;;
  list)
    list_ns
    exit $?
    ;;
  help)
    cmd_usage
    exit $?
    ;;
  "")
    echo "Please provide a command"
    cmd_usage
    exit 1
    ;;
  *)
    echo "Unknown command: $1"
    cmd_usage
    exit 1
esac
