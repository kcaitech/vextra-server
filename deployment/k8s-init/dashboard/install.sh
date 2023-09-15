#!/bin/bash

# 安装dashboard

set -e

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

# 获取apiserver_ip
echo "请输入apiserver_ip"
read -r -p "apiserver_ip（同一网段可只输入最后一个数字）：" apiserver_ip
if [[ "$apiserver_ip" == "" ]]; then
  echo "输入错误"
  exit 1
fi
# 若只输入了ip的最后一个数字，则使用this_ip的前三个数字拼接
if [[ "$apiserver_ip" =~ ^[0-9]+$ ]]; then
  apiserver_ip="${this_ip%.*}.$apiserver_ip"
fi

# 下载配置文件 dashboard
echo "下载配置文件 dashboard"
curl https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml -L -o kubernetes-dashboard.yaml

echo "安装dashboard"
awk '
BEGIN { flag_kind=0; flag_name=0; flag_target_port=0; }
/king: / { flag_kind=0; flag_name=0; flag_target_port=0; }
/kind: Service$/ { flag_kind=1; }
flag_kind && /name: kubernetes-dashboard$/ { flag_name=1; }
flag_name && /targetPort: 8443/ { flag_target_port=1; }
{
  print; # 默认打印当前行
  if (flag_target_port) {
    spaces = $0;
    sub(/targetPort:.*/, "", spaces); # 获取行首的空格
    print spaces "nodePort: 30001";
    sub(/^.{4}/, "", spaces); # 减4个空格
    print spaces "type: NodePort";
    flag_kind=0; flag_name=0; flag_target_port=0;
  }
}' kubernetes-dashboard.yaml > kubernetes-dashboard.yaml.tmp && mv kubernetes-dashboard.yaml.tmp kubernetes-dashboard.yaml
kubectl apply -f kubernetes-dashboard.yaml
kubectl create serviceaccount dashboard-admin -n kubernetes-dashboard
kubectl create clusterrolebinding dashboard-admin --clusterrole=cluster-admin --serviceaccount=kubernetes-dashboard:dashboard-admin
kubectl create token dashboard-admin -n kubernetes-dashboard
echo "请在浏览器中打开：https://$this_ip:30001或https://$apiserver_ip:30001"
echo "并将dashboard-admin-token.log中的token复制到登录页面中，或者执行以下命令重新获取token："
echo "kubectl create token dashboard-admin -n kubernetes-dashboard"
