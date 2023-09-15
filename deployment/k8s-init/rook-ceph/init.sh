#!/bin/bash

# 在每个要作为osd的节点运行
# 运行本脚本前：
# 1、确保要作为osd的节点已挂载未格式化的硬盘，可运行lsblk -f、fdisk -l等命令查看

set -e

apt update
apt install -y lvm2 gdisk

if ! grep -q "^rbd" /etc/modules; then
  echo "rbd" >> /etc/modules
fi
modprobe rbd

sed -i 's/LimitNOFILE=infinity/LimitNOFILE=1048576/g' /lib/systemd/system/containerd.service
systemctl daemon-reload
systemctl restart containerd
