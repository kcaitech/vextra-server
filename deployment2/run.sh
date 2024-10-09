#!/bin/bash

dirs=("db" "server" "entry" "https")

cmd=$1

if [ "$cmd" != "up" ] && [ "$cmd" != "down" ]; then
    echo "Usage: $0 [ up | down ]"
    exit 1
fi

for d in "${dirs[@]}"
do
    echo "--- starting $d ---"
    pushd $d 1>/dev/null
    ./run.sh $cmd
    popd 1>/dev/null
done
