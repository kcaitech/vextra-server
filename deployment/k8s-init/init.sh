#!/bin/bash

# 获取用于haproxy的所有name、ip和端口，多个之间以,隔开 格式：name ip:port
echo "请输入其余节点的hostname和ip，多个之间以,隔开 格式：name ip"
echo "（kc-master2 192.168.137.21,kc-master3 192.168.137.22）"
read -r master_nodes
# 验证格式以及分割
if [[ "$master_nodes" == "" ]]; then
  master_nodes="kc-master2 192.168.137.21,kc-master3 192.168.137.22"
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
