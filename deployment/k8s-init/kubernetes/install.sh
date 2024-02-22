#!/bin/bash

# 执行本脚本前请先设置好以下参数：
# 1. 主机名
# hostnamectl set-hostname my-hostname
# bash
# 2. 静态ip（若已在路由器或交换机中指定了本机IP则可跳过此步）
# vim /etc/netplan/00-installer-config.yaml
# netplan apply
# 3. root用户（需以root用户执行本脚本）
# passwd root

# 参考文档：
# https://kubernetes.io/zh-cn/docs/reference/config-api/kubeadm-config.v1beta3

# 运行环境：Ubuntu 22.04.3 LTS

set -e

echo "请选择初始化类型"
echo "1) 集群初始化"
echo "2) master节点初始化"
echo "3) worker节点初始化"
read -r init_type
if [[ "$init_type" != "1" && "$init_type" != "2" && "$init_type" != "3" ]]; then
  echo "输入错误"
  exit 1
fi

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

node_name=$(hostname)
read -r -p "请输入节点名称（$node_name）" node_name
if [[ "$node_name" == "" ]]; then
  node_name=$(hostname)
fi

# 获取用于haproxy的所有name、ip和端口，多个之间以,隔开 格式：name ip:port
echo "请输入所有master节点的name、ip和端口，多个之间以,隔开 格式：name ip:port"
echo "（kc-master1 172.16.0.20:6443,kc-master2 172.16.0.21:6443,kc-master3 172.16.0.22:6443）"
read -r master_nodes
other_master_nodes_str=""
# 验证格式以及分割
if [[ "$master_nodes" == "" ]]; then
  master_nodes="kc-master1 172.16.0.20:6443,kc-master2 172.16.0.21:6443,kc-master3 172.16.0.22:6443"
fi
IFS=',' read -ra master_nodes <<< "$master_nodes"
for node in "${master_nodes[@]}"; do
  if [[ ! "$node" =~ ^[a-zA-Z0-9_-]+[[:space:]]+[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$ ]]; then
    echo "输入错误"
    exit 1
  fi
  ip_port=$(echo "$node" | awk '{print $2}')
  ip=$(echo "$ip_port" | cut -d: -f1)
  # 如果ip不等于this_ip则加入other_master_nodes_str
  if [[ "$ip" != "$this_ip" ]]; then
    other_master_nodes_str+="    $ip\n"
  fi
done

# 获取keepalived的vip和state
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  echo "请输入keepalived的vip和state（MASTER/BACKUP）"
  read -r -p "vip（同一网段可只输入最后一个数字）（19）：" keepalived_vip
  if [[ "$keepalived_vip" == "" ]]; then
    keepalived_vip="19"
  fi
  # 若只输入了vip的最后一个数字，则使用this_ip的前三个数字拼接
  if [[ "$keepalived_vip" =~ ^[0-9]+$ ]]; then
    keepalived_vip="${this_ip%.*}.$keepalived_vip"
  fi
  read -r -p "state（MASTER）：" keepalived_state
  if [[ "$keepalived_state" == "" ]]; then
    keepalived_state="MASTER"
  fi
  if [[ "$keepalived_state" != "MASTER" && "$keepalived_state" != "BACKUP" ]]; then
    echo "输入错误"
    exit 1
  fi
fi

# 获取ApiServer的域名、IP和端口
echo "请输入ApiServer的endpoint信息"
read -r -p "域名（cluster-endpoint）：" apiserver_domain
if [[ "$apiserver_domain" == "" ]]; then
  apiserver_domain="cluster-endpoint"
fi
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  apiserver_ip=$keepalived_vip
else
  # 获取ApiServer的IP
  read -r -p "IP（同一网段可只输入最后一个数字）：" apiserver_ip
  if [[ "$apiserver_ip" == "" ]]; then
    echo "输入错误"
    exit 1
  fi
  if [[ "$apiserver_ip" =~ ^[0-9]+$ ]]; then
    apiserver_ip="${this_ip%.*}.$apiserver_ip"
  fi
fi
# 写入hosts
echo "写入hosts $apiserver_ip $apiserver_domain"
echo "$apiserver_ip $apiserver_domain" >> /etc/hosts
# 获取端口
read -r -p "端口（9443）：" apiserver_port
if [[ "$apiserver_port" == "" ]]; then
  apiserver_port="9443"
fi

# 获取docker-registry的域名和IP
echo "请输入docker-registry的endpoint信息"
read -r -p "域名（registry.protodesign.cn）：" registry_domain
if [[ "$registry_domain" == "" ]]; then
  registry_domain="registry.protodesign.cn"
fi
read -r -p "IP（同一网段可只输入最后一个数字）（121.199.25.192）：" registry_ip
if [[ "$registry_ip" == "" ]]; then
    registry_ip="121.199.25.192"
fi
if [[ "$registry_ip" =~ ^[0-9]+$ ]]; then
  registry_ip="${this_ip%.*}.$registry_ip"
fi
# 写入hosts
echo "写入hosts $registry_ip $registry_domain"
echo "$registry_ip $registry_domain" >> /etc/hosts

# 获取join时需要的token
if [[ "$init_type" == "2" || "$init_type" == "3" ]]; then
  echo "请输入join token（可在集群上执行以下命令获取: kubeadm token create --print-join-command | awk -F'--token ' '{print \$2}' | awk '{print \$1}'）"
  read -r join_token
  if [[ "$join_token" == "" ]]; then
    echo "输入错误"
    exit 1
  fi
fi

# 获取certificateKey
if [[ "$init_type" == "2" ]]; then
  echo "请输入certificateKey（可在现有的master节点上执行kubeadm init phase upload-certs --upload-certs获取）"
  read -r certificate_key
  if [[ "$certificate_key" == "" ]]; then
    echo "输入错误"
    exit 1
  fi
fi

# 安装基础软件
echo "安装基础软件"
apt update
sleep 3
apt install -y wget ca-certificates curl gnupg htop git jq tree

# 安装docker
./install-docker-ubuntu.sh

# 设置containerd参数
echo "设置containerd参数"
containerd config default > /etc/containerd/config.toml
sed -i 's#sandbox_image\s*=\s*".*"#sandbox_image = "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.9"#' /etc/containerd/config.toml
sed -i 's/SystemdCgroup\s*=\s*false/SystemdCgroup = true/' /etc/containerd/config.toml
awk '/\[plugins\."io\.containerd\.grpc\.v1\.cri"\.registry\.mirrors\]/ {
  print;
  indent = match($0, /[^ \t]/) - 1; # 找到第一个非空白字符的位置
  prefix = substr($0, 1, indent);  # 提取行前的空白字符
  print prefix "  [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker.io\"]";
  print prefix "    endpoint = [\"https://jsoixv4u.mirror.aliyuncs.com\", \"https://registry-1.docker.io\"]";
  print prefix "  [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"registry.protodesign.cn:36000\"]";
  print prefix "    endpoint = [\"http://registry.protodesign.cn:36000\"]";
  next;
}
1' /etc/containerd/config.toml > /etc/containerd/config.toml.tmp && mv /etc/containerd/config.toml.tmp /etc/containerd/config.toml
sed -i 's/LimitNOFILE=infinity/LimitNOFILE=1048576/g' /lib/systemd/system/containerd.service
echo "重启containerd服务"
systemctl daemon-reload
systemctl enable --now containerd
systemctl restart containerd

# 配置haproxy
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  echo "配置haproxy"
  cp haproxy.template.cfg haproxy.cfg
  sed -i "s/\$apiserver_port/$apiserver_port/g" haproxy.cfg
  for node in "${master_nodes[@]}"; do
    name=$(echo "$node" | awk '{print $1}')
    ip_port=$(echo "$node" | awk '{print $2}')
    echo "  server $name $ip_port check" >> haproxy.cfg
  done
  mkdir /usr/local/k8s-init/haproxy -p
  mv haproxy.cfg /usr/local/k8s-init/haproxy/haproxy.cfg
  # 以docker容器方式运行haproxy
  echo "以docker容器方式运行haproxy"
  docker run -d --name haproxy \
  --net=host \
  --restart=always \
  -v /usr/local/k8s-init/haproxy/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro \
  haproxytech/haproxy-ubuntu:2.8
fi

# 配置keepalived
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  echo "配置keepalived"
  cp keepalived.template.conf keepalived.conf
  sed -i "s/\$node_name/$node_name/g" keepalived.conf
  sed -i "s/\$keepalived_state/$keepalived_state/g" keepalived.conf
  sed -i "s/\$net_card_name/$net_card_name/g" keepalived.conf
  sed -i "s/\$keepalived_vip/$keepalived_vip/g" keepalived.conf
  sed -i "s/\$node_ip/$this_ip/g" keepalived.conf
  sed -i "s/    \$other_node_ip/$other_master_nodes_str/g" keepalived.conf
  if [[ "$init_type" == "1" ]]; then
    sed -i "s/\$priority/100/g" keepalived.conf
  else
    sed -i "s/\$priority/80/g" keepalived.conf
  fi
  mkdir /usr/local/k8s-init/keepalived -p
  mv keepalived.conf /usr/local/k8s-init/keepalived/keepalived.conf
  # 配置check_haproxy.sh
  echo "配置check_haproxy.sh"
  cp check_haproxy.sh /usr/local/k8s-init/keepalived/check_haproxy.sh
  chmod +x /usr/local/k8s-init/keepalived/check_haproxy.sh
  # 以docker容器方式运行keepalived
  echo "以docker容器方式运行keepalived"
  docker run -d --name keepalived \
  --net=host \
  --cap-add=NET_ADMIN \
  --cap-add=NET_BROADCAST \
  --cap-add=NET_RAW \
  --restart=always \
  -v /usr/local/k8s-init/keepalived/keepalived.conf:/container/service/keepalived/assets/keepalived.conf \
  -v /usr/local/k8s-init/keepalived/check_haproxy.sh:/usr/bin/check_haproxy.sh \
  osixia/keepalived:2.0.20 --copy-service
fi

# 安装kubernetes相关组件
echo "安装kubernetes相关组件"
curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
echo "deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list
apt update
apt install -y kubelet=1.28.1-00 kubeadm=1.28.1-00
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  if ! dpkg -l | grep -q kubectl; then
    apt install -y kubectl=1.28.1-00
  fi
fi

# 创建集群
if [[ "$init_type" == "1" ]]; then
  # 设置kubernetes参数
  echo "设置kubernetes参数"
  cp k8s-init.template.yaml k8s-init.yaml
  sed -i "s/\$this_ip/$this_ip/g" k8s-init.yaml
  sed -i "s/\$node_name/$node_name/g" k8s-init.yaml
  sed -i "s/\$apiserver_domain/$apiserver_domain/g" k8s-init.yaml
  sed -i "s/\$apiserver_port/$apiserver_port/g" k8s-init.yaml
  sed -i "s/\$apiserver_ip/$apiserver_ip/g" k8s-init.yaml
  for node in "${master_nodes[@]}"; do
    name=$(echo "$node" | awk '{print $1}')
    ip_port=$(echo "$node" | awk '{print $2}')
    ip=$(echo "$ip_port" | cut -d: -f1)
    echo "    - $name" >> k8s-init.yaml
    echo "    - $ip" >> k8s-init.yaml
  done
  # 手动修改参数：name、advertiseAddress等
  read -r -p "是否需要手动修改kubeadm init config参数（y/n）" init_type
  if [[ "$init_type" == "" || "$init_type" == "y" ]]; then
      vim k8s-init.yaml
  fi

  # 初始化kubernetes集群
  echo "初始化kubernetes集群"
  kubeadm init --config k8s-init.yaml --upload-certs | tee kubeadm-init.log
  mkdir -p $HOME/.kube
  cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  chown "$(id -u):$(id -g)" $HOME/.kube/config

  # 安装网络插件calico
  echo "安装网络插件calico"
  kubectl apply -f calico.yaml

  # 清除master节点污点
  echo "清除master节点污点"
  kubectl taint nodes --all node-role.kubernetes.io/master-
fi

# 加入集群
if [[ "$init_type" == "2" || "$init_type" == "3" ]]; then
  # 设置JoinConfiguration参数
  echo "设置JoinConfiguration参数"
  cp node-join.template.yaml node-join.yaml
  sed -i "s/\$join_token/$join_token/g" node-join.yaml
  sed -i "s/\$apiserver_domain/$apiserver_domain/g" node-join.yaml
  sed -i "s/\$apiserver_port/$apiserver_port/g" node-join.yaml
  sed -i "s/\$node_name/$node_name/g" node-join.yaml
  if [ "$init_type" == "2" ]; then # 加入master节点
  cat <<EOF >> node-join.yaml
controlPlane:
  localAPIEndpoint:
    advertiseAddress: $this_ip
    bindPort: 6443
  certificateKey: $certificate_key
EOF
  fi
  # 加入集群
  echo "加入集群"
  kubeadm join --config node-join.yaml
  # 设置kubectl
  if [[ "$init_type" == "2" ]]; then
    mkdir -p $HOME/.kube
    cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    chown "$(id -u):$(id -g)" $HOME/.kube/config
    # 清除master节点污点
    echo "清除master节点污点"
    kubectl taint nodes --all node-role.kubernetes.io/master-

#    # 修改kubelet参数
#    echo "修改kubelet参数（等待30秒）"
#    sleep 30
#    formatted_date=$(date +"%Y%m%d_%H%M%S")_$(date +%N | cut -c1-6)
#    cp /var/lib/kubelet/config.yaml /var/lib/kubelet/config.yaml.bak.$formatted_date
#    awk '
#    BEGIN { flag_system_reserved=0; flag_memory=0; }
#    /systemReserved:/ { flag_system_reserved=1; }
#    flag_system_reserved && /memory:/ { flag_memory=1; }
#    {
#      if (!flag_memory) {
#        print; # 默认打印当前行
#      } else {
#        spaces = $0;
#        sub(/memory:.*/, "", spaces); # 获取行首的空格
#        print spaces "memory: 256Mi";
#        flag_system_reserved=0; flag_memory=0;
#      }
#    }' /var/lib/kubelet/config.yaml > /var/lib/kubelet/config.yaml.tmp && mv /var/lib/kubelet/config.yaml.tmp /var/lib/kubelet/config.yaml
#    # 重启kubelet
#    echo "重启kubelet"
#    systemctl daemon-reload
#    systemctl restart kubelet

  fi
fi
