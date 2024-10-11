#!/bin/bash

dirs=("db" "server" "entry" "https")

cmd=$1

if [ "$cmd" != "up" ] && [ "$cmd" != "down" ] && [ "$cmd" != "reset" ]; then
    echo "Usage: $0 [ up | down | reset ]"
    exit 1
fi
if [ "$cmd" = "up" ]; then
    msg="starting"
elif [ "$cmd" = "down" ]; then
    msg="stopping"
else
    msg="reseting"
fi
for d in "${dirs[@]}"
do
    echo "--- $msg $d ---"
    pushd $d 1>/dev/null
    ./run.sh $cmd
    popd 1>/dev/null
done
