#!/bin/bash

# 安装ceph
# 运行本脚本前要先：
#   1、将其余两台主机添加信任密钥，具体查看../kubernetes/readme
#   2、将其余两台主机添加到/etc/hosts，运行../kubernetes/install.sh后即已添加
# 本脚本仅在第一个master节点运行即可

# 获取网卡名称
read -r -p "请输入网卡名称（eth0）" net_card_name
if [[ "$net_card_name" == "" ]]; then
  net_card_name="eth0"
fi

# 获取本机ip
this_ip=$(ip addr show $net_card_name | grep "inet\b" | awk '{print $2}' | cut -d/ -f1) # 网卡下的ip
if [[ "$this_ip" == "" ]]; then
  echo "获取网卡ip错误，请检查网卡名称（$net_card_name）是否正确"
  exit 1
fi
gateway_ip="${this_ip%.*}.1" # 网卡下的网关ip

# 获取网络代理地址
echo "请输入网络代理地址（http、socks5）（包含协议、ip和端口），不设置代理请输入空格"
read -r -p "（http://$gateway_ip:10809）" proxy_address
if [[ "$proxy_address" == "" ]]; then
  proxy_address="http://$gateway_ip:10809"
elif [[ "$proxy_address" == " " ]]; then
  proxy_address=""
fi

echo "请输入其余两个节点的hostname，多个之间以,隔开"
echo "（kc-master2,kc-master3）"
read -r master_nodes
# 验证格式以及分割
if [[ "$master_nodes" == "" ]]; then
  master_nodes="kc-master2,kc-master3"
fi
IFS=',' read -ra master_nodes <<< "$master_nodes"
for node in "${master_nodes[@]}"; do
  if [[ ! "$node" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    echo "输入错误"
    exit 1
  fi
done

# 安装cephadm
# 参考：https://docs.ceph.com/en/latest/cephadm/install/
apt install -y cephadm
# 创建ceph集群
cephadm bootstrap --mon-ip $this_ip

# 安装ceph-cli（ceph-common）
export http_proxy=$proxy_address
export https_proxy=$proxy_address
export HTTP_PROXY=$proxy_address
export HTTPS_PROXY=$proxy_address
cephadm add-repo --release reef
cephadm install ceph-common
export HTTP_PROXY=
export HTTPS_PROXY=
export http_proxy=
export https_proxy=

# 添加主机
# 参考：https://docs.ceph.com/en/latest/cephadm/host-management/
for node in "${master_nodes[@]}"; do
  ssh-copy-id -f -i /etc/ceph/ceph.pub root@$node
  ceph orch host add $node --labels _admin
done

# 添加osd
# 参考：https://docs.ceph.com/en/latest/cephadm/services/osd/#cephadm-deploy-osds
# 使用任何可用和未使用的存储设备：ceph orch apply osd --all-available-devices
for node in "${master_nodes[@]}"; do
  ceph orch daemon add osd $node:/dev/sdb
done

# 创建文件系统
# 参考：https://docs.ceph.com/en/latest/cephfs/#getting-started-with-cephfs
ceph fs volume create cephfs
