#!/bin/bash

# 在其中一个master节点执行
# 运行本脚本前：
# 1、确保要作为osd的节点已挂载未格式化的硬盘，可运行lsblk -f、fdisk -l等命令查看
# 2、修改operator.yaml中的一些镜像地址：data.ROOK_CSI_CEPH_IMAGE等
# 3、修改cluster.yaml中的spec.storage.nodes

set -e


# 获取用于ceph的所有node的名称，多个之间以,隔开
echo "请输入获取用于ceph的所有node的名称，多个之间以,隔开"
echo "（kc-master1,kc-master2,kc-master3）"
read -r ceph_nodes
# 验证格式以及分割
if [[ "$ceph_nodes" == "" ]]; then
#  echo "输入错误"
#  exit 1
  ceph_nodes="kc-master1,kc-master2,kc-master3"
fi
IFS=',' read -ra ceph_nodes <<< "$ceph_nodes"
if [[ ${#ceph_nodes[@]} -lt 3 ]]; then
  echo "请至少输入3个节点"
  exit 1
fi
for node in "${ceph_nodes[@]}"; do
  if [[ ! "$node" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    echo "输入错误"
    exit 1
  fi
  # 给node打标签
  kubectl label node $node ceph-role="true"
done

apt update
apt install -y lvm2 gdisk

if ! grep -q "^rbd" /etc/modules; then
  echo "rbd" >> /etc/modules
fi
modprobe rbd

sed -i 's/LimitNOFILE=infinity/LimitNOFILE=1048576/g' /lib/systemd/system/containerd.service
systemctl daemon-reload
systemctl restart containerd

helm repo add rook-release https://charts.rook.io/release
helm repo update
helm install --create-namespace --namespace rook-ceph rook-ceph rook-release/rook-ceph -f cluster-values.yaml
