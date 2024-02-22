#!/bin/bash

set -e

# 赋予所有脚本执行权限
find . -type f -name "*.sh" -exec chmod +x {} \;

# 将其余节点的hostname和ip加入到hosts文件，多个之间以,隔开 格式：hostname ip
echo "请输入所有master节点的hostname和ip，多个之间以,隔开 格式：name ip"
echo "（kc-master1 172.16.0.20,kc-master2 172.16.0.21,kc-master3 172.16.0.22）"
read -r master_nodes
# 验证格式以及分割
if [[ "$master_nodes" == "" ]]; then
  master_nodes="kc-master1 172.16.0.20,kc-master2 172.16.0.21,kc-master3 172.16.0.22"
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
apt update

# 设置时区
echo "设置时区"
apt install -y systemd-timesyncd
systemctl enable systemd-timesyncd
systemctl start systemd-timesyncd
timedatectl set-timezone Asia/Shanghai
echo "NTP=ntp.tencent.com" >> /etc/systemd/timesyncd.conf
echo "FallbackNTP=ntp1.tencent.com,ntp2.tencent.com,ntp3.tencent.com" >> /etc/systemd/timesyncd.conf
echo "RootDistanceMaxSec=5" >> /etc/systemd/timesyncd.conf
echo "PollIntervalMinSec=32" >> /etc/systemd/timesyncd.conf
echo "PollIntervalMaxSec=2048" >> /etc/systemd/timesyncd.conf
systemctl restart systemd-timesyncd
timedatectl set-ntp on

# 安装IO监控工具
apt install -y iotop

# 禁用swap
echo "禁用swap"
cp /etc/fstab /etc/fstab.bak.$formatted_date
sed -i '/[ \t]swap[ \t]/ s/^/# /' /etc/fstab
swapoff -a

# 设置ssh参数
echo "设置ssh参数"
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.bak.$formatted_date
awk '
/ClientAliveInterval [0-9]+/ {
  print "ClientAliveInterval 60";
  next
}
/ClientAliveCountMax [0-9]+/ {
  print "ClientAliveCountMax 60";
  next
}
{
  print
}
' /etc/ssh/sshd_config > /etc/ssh/sshd_config.tmp && mv /etc/ssh/sshd_config.tmp /etc/ssh/sshd_config
service ssh restart

# 关闭防火墙
echo "关闭防火墙"
systemctl stop ufw
systemctl disable ufw

# 清空iptables
echo "清空iptables"
iptables -F
iptables -X
iptables -Z
iptables -t nat -F
iptables -t nat -X
iptables -t nat -Z
iptables -t mangle -F
iptables -t mangle -X
iptables -t mangle -Z
iptables -t raw -F
iptables -t raw -X
iptables -t raw -Z
