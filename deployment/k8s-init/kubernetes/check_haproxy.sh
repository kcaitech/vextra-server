#!/bin/bash

count=`netstat -apn | grep 9443 | grep haproxy | wc -l`
if [ $count -gt 0 ]; then
  exit 0
else
  exit 1
fi
