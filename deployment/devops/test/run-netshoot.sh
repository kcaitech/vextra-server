#!/bin/bash

set -e

formatted_date=$(date +"%Y%m%d-%H%M%S")-$(date +%N | cut -c1-6)
kubectl run -i --tty --rm --restart=Never \
  "test-netshoot-$formatted_date" \
  --image=docker.io/nicolaka/netshoot:latest "$*" \
  -- bash
