#!/bin/bash

# 安装kuboard

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

# 安装运行
docker run -d \
  --restart=unless-stopped \
  --name=kuboard \
  -p 30002:80/tcp \
  -p 10081:10081/tcp \
  -e KUBOARD_ENDPOINT="http://$this_ip:30002" \
  -e KUBOARD_AGENT_SERVER_TCP_PORT="10081" \
  -v /root/kuboard-data:/data \
  swr.cn-east-2.myhuaweicloud.com/kuboard/kuboard:v3

# 输出信息
echo "http://$this_ip:30002"
echo "初始账号密码：admin Kuboard123"
