#!/bin/bash

# 安装mysql

# 获取网卡名称
read -r -p "请输入网卡名称（eth0）" net_card_name
if [[ "$net_card_name" == "" ]]; then
  net_card_name="eth0"
fi
# 获取本机ip
this_ip=$(ifconfig $net_card_name | grep 'inet ' | awk '{print $2}' | cut -d':' -f2) # 网卡下的ip
if [[ "$this_ip" == "" ]]; then
  echo "获取网卡ip错误，请检查网卡名称（$net_card_name）是否正确"
  exit 1
fi
gateway_ip="${this_ip%.*}.1" # 网卡下的网关ip

echo "请输入网络代理地址（http、socks5）（包含协议、ip和端口），不设置代理请输入空格"
read -r -p "（http://$gateway_ip:10809）" proxy_address
if [[ "$proxy_address" == "" ]]; then
  proxy_address="http://$gateway_ip:10809"
elif [[ "$proxy_address" == " " ]]; then
  proxy_address=""
fi

# 设置代理
export http_proxy=$proxy_address
export https_proxy=$proxy_address
export HTTP_PROXY=$proxy_address
export HTTPS_PROXY=$proxy_address
export no_proxy=localhost,127.0.0.1,cluster-endpoint
export NO_PROXY=localhost,127.0.0.1,cluster-endpoint
# 添加仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
# 安装
helm install bitnami bitnami/mysql -n mysql --create-namespace -f values.yaml
# 取消代理
export HTTP_PROXY=
export HTTPS_PROXY=
export http_proxy=
export https_proxy=
export no_proxy=
export NO_PROXY=
