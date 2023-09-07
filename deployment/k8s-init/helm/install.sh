#!/bin/bash

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

echo "请输入网络代理地址（http、socks5）（包含协议、ip和端口），不设置代理请输入空格"
read -r -p "（http://$gateway_ip:10809）" proxy_address
if [[ "$proxy_address" == "" ]]; then
  proxy_address="http://$gateway_ip:10809"
elif [[ "$proxy_address" == " " ]]; then
  proxy_address=""
fi

# 下载安装包
echo "下载安装包"
export http_proxy="${proxy_address}"
export https_proxy="${proxy_address}"
export HTTP_PROXY="${proxy_address}"
export HTTPS_PROXY="${proxy_address}"
curl https://get.helm.sh/helm-v3.12.3-linux-amd64.tar.gz -LO
export HTTP_PROXY=
export HTTPS_PROXY=
export http_proxy=
export https_proxy=
# 解压并复制到/usr/local/bin
mkdir helm-v3.12.3-linux-amd64
tar -zxvf helm-v3.12.3-linux-amd64.tar.gz -C helm-v3.12.3-linux-amd64
cp helm-v3.12.3-linux-amd64/linux-amd64/helm /usr/local/bin/helm
# 添加repo
helm repo add stable  https://kubernetes.oss-cn-hangzhou.aliyuncs.com/charts
helm repo update
