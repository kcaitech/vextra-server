#!/bin/bash

set -e

# 赋予所有脚本执行权限
find . -type f -name "*.sh" -exec chmod +x {} \;

# 将其余节点的hostname和ip加入到hosts文件，多个之间以,隔开 格式：hostname ip
echo "请输入所有master节点的hostname和ip，多个之间以,隔开 格式：name ip"
echo "（kc-master1 192.168.137.20,kc-master2 192.168.137.21,kc-master3 192.168.137.22）"
read -r master_nodes
# 验证格式以及分割
if [[ "$master_nodes" == "" ]]; then
  master_nodes="kc-master1 192.168.137.20,kc-master2 192.168.137.21,kc-master3 192.168.137.22"
fi
IFS=',' read -ra master_nodes <<< "$master_nodes"
for node in "${master_nodes[@]}"; do
  if [[ "$node" =~ ^[a-zA-Z0-9_-]+[[:space:]]+[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    # 写入hosts
    name=$(echo "$node" | awk '{print $1}')
    ip=$(echo "$node" | awk '{print $2}')
    echo "写入hosts $ip $name"
    echo "$ip $name" >> /etc/hosts
  else
    echo "输入错误"
    exit 1
  fi
done

# 设置时区
echo "设置时区"
timedatectl set-timezone Asia/Shanghai
echo "NTP=cg.lzu.edu.cn" >> /etc/systemd/timesyncd.conf
systemctl restart systemd-timesyncd

# 设置内核参数
echo "设置内核参数"
echo "overlay" >> /etc/modules-load.d/k8s.conf
echo "br_netfilter" >> /etc/modules-load.d/k8s.conf
modprobe overlay
modprobe br_netfilter
cp etc_sysctl.d_k8s.conf /etc/sysctl.d/k8s.conf
cat etc_security_limits.conf >> /etc/security/limits.conf
sysctl -p
sysctl --system

# 设置apt源
echo "设置apt源"
formatted_date=$(date +"%Y%m%d_%H%M%S")_$(date +%N | cut -c1-6)
cp /etc/apt/sources.list /etc/apt/sources.list.bak.$formatted_date
sed -i 's@//.*archive.ubuntu.com@//mirrors.ustc.edu.cn@g' /etc/apt/sources.list

# 安装IO监控工具
apt update
apt install -y iotop

# 禁用swap
echo "禁用swap"
cp /etc/fstab /etc/fstab.bak.$formatted_date
sed -i '/[ \t]swap[ \t]/ s/^/# /' /etc/fstab
swapoff -a

