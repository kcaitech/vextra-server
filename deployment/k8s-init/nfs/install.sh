#!/bin/bash

set -e

apt update && apt install -y nfs-kernel-server
cat << EOF > /etc/exports
/root/nfs_root/ *(insecure,rw,sync,no_root_squash)
EOF
mkdir /root/nfs_root

# 客户端执行
# apt update && apt install -y nfs-common
# 客户端测试是否成功
# mkdir /tmp/testnfs && mount -t nfs 192.168.137.20:/root/nfs_root /tmp/testnfs && echo "hello nfs" >> /tmp/testnfs/test.txt && cat /tmp/testnfs/test.txt
