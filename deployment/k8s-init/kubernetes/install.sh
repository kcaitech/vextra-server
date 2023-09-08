#!/bin/bash

# 执行本脚本前请先设置好以下参数：
# 1. 主机名
# hostnamectl set-hostname my-hostname
# bash
# 2. 静态ip（若已在路由器或交换机中指定了本机IP则可跳过此步）
# vim /etc/netplan/00-installer-config.yaml
# netplan apply
# 3. root用户（需以root用户执行本脚本）
# sudo passwd root

# 参考文档：
# https://kubernetes.io/zh-cn/docs/reference/config-api/kubeadm-config.v1beta3

# 运行环境：Ubuntu 22.04.3 LTS

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
echo "（kc-master1 192.168.137.20:6443,kc-master2 192.168.137.21:6443,kc-master3 192.168.137.22:6443）"
read -r master_nodes
# 验证格式以及分割
if [[ "$master_nodes" == "" ]]; then
  master_nodes="kc-master1 192.168.137.20:6443,kc-master2 192.168.137.21:6443,kc-master3 192.168.137.22:6443"
fi
IFS=',' read -ra master_nodes <<< "$master_nodes"
for node in "${master_nodes[@]}"; do
  if [[ ! "$node" =~ ^[a-zA-Z0-9_-]+[[:space:]]+[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$ ]]; then
    echo "输入错误"
    exit 1
  fi
done

# 获取keepalived的vip和state
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  echo "请输入keepalived的vip和state（MASTER/BACKUP）"
  read -r -p "vip（同一网段可只输入最后一个数字）：" keepalived_vip
  if [[ "$keepalived_vip" == "" ]]; then
    echo "输入错误"
    exit 1
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
echo "写入hosts $apiserver_domain $apiserver_ip"
echo "$apiserver_ip $apiserver_domain" >> /etc/hosts
read -r -p "端口（9443）：" apiserver_port
if [[ "$apiserver_port" == "" ]]; then
  apiserver_port="9443"
fi

# 获取join时需要的token
if [[ "$init_type" == "2" || "$init_type" == "3" ]]; then
  echo "请输入join token（可在集群上执行kubeadm token create --print-join-command获取）"
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

# 获取网络代理地址
echo "请输入网络代理地址（http、socks5）（包含协议、ip和端口），不设置代理请输入空格"
read -r -p "（http://$gateway_ip:10809）" proxy_address
if [[ "$proxy_address" == "" ]]; then
  proxy_address="http://$gateway_ip:10809"
elif [[ "$proxy_address" == " " ]]; then
  proxy_address=""
fi

# 设置时区
echo "设置时区"
timedatectl set-timezone Asia/Shanghai
echo "NTP=cg.lzu.edu.cn" >> /etc/systemd/timesyncd.conf
systemctl restart systemd-timesyncd

# 设置内核参数
echo "设置内核参数"
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF
modprobe overlay
modprobe br_netfilter
cat << EOF > /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1

fs.may_detach_mounts = 1
vm.overcommit_memory=1
vm.panic_on_oom=0
fs.inotify.max_user_watches=89100
fs.file-max=52706963
fs.nr_open=52706963
net.netfilter.nf_conntrack_max=2310720

net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_probes = 3
net.ipv4.tcp_keepalive_intvl =15
net.ipv4.tcp_max_tw_buckets = 36000
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_max_orphans = 327680
net.ipv4.tcp_orphan_retries = 3
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 16384
net.ipv4.ip_conntrack_max = 65536
net.ipv4.tcp_max_syn_backlog = 16384
net.ipv4.tcp_timestamps = 0
net.core.somaxconn = 16384

net.ipv6.conf.all.disable_ipv6 = 0
net.ipv6.conf.default.disable_ipv6 = 0
net.ipv6.conf.lo.disable_ipv6 = 0
net.ipv6.conf.all.forwarding = 1
EOF
cat << EOF >> /etc/security/limits.conf
* soft nproc 102400
* hard nproc 104800
* soft nofile 102400
* hard nofile 104800
root soft nproc 102400
root hard nproc 104800
root soft nofile 102400
root hard nofile 104800
EOF
sysctl -p
sysctl --system

# 设置apt源
echo "设置apt源"
formatted_date=$(date +"%Y%m%d%H%M%S%3N")
cp /etc/apt/sources.list /etc/apt/sources.list.bak.$formatted_date
sed -i 's@//.*archive.ubuntu.com@//mirrors.ustc.edu.cn@g' /etc/apt/sources.list

# 安装基础软件
echo "安装基础软件"
apt-get update
apt-get install -y sudo wget ca-certificates curl gnupg htop git jq tree

# 安装docker
echo "安装docker"
rm -f /etc/apt/keyrings/docker.gpg
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://mirrors.aliyun.com/docker-ce/linux/ubuntu \
$(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin docker-compose
# 设置docker参数
echo "设置docker参数"
cat << EOF > /etc/docker/daemon.json
{
    "exec-opts": ["native.cgroupdriver=systemd"],
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m"
    },
    "storage-driver": "overlay2",
    "registry-mirrors": [
      "https://jsoixv4u.mirror.aliyuncs.com",
      "https://docker.mirrors.ustc.edu.cn",
      "http://hub-mirror.c.163.com"
    ]
}
EOF
# 重启docker服务
echo "重启docker服务"
systemctl enable --now docker
systemctl restart docker
# 验证docker服务是否正常
#echo "验证docker服务是否正常"
#docker images
#docker pull busybox
#docker run --rm busybox uname -m

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
  next;
}
1' /etc/containerd/config.toml > /etc/containerd/config.toml.tmp && mv /etc/containerd/config.toml.tmp /etc/containerd/config.toml
sed -i 's/LimitNOFILE=infinity/LimitNOFILE=1048576/g' /lib/systemd/system/containerd.service
systemctl daemon-reload
systemctl enable --now containerd
systemctl restart containerd

# 配置haproxy
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  echo "配置haproxy"
  cat <<EOF > haproxy.cfg
global
  log 127.0.0.1 local2
  chroot /var/lib/haproxy
  pidfile /var/run/haproxy.pid
  maxconn 4096
  stats socket /var/lib/haproxy/stats.sock mode 660 level admin expose-fd listeners
  stats timeout 30s
  user haproxy
  group haproxy
  daemon

defaults
  log global
  mode http
  option httplog
  option dontlognull
  option http-server-close
  option forwardfor except 127.0.0.0/8
  option redispatch
  retries 3
  timeout http-request 10s
  timeout queue 1m
  timeout connect 10s
  timeout client 1m
  timeout server 1m
  timeout http-keep-alive 10s
  timeout check 10s
  maxconn 3000

frontend kube-apiserver
  bind *:$apiserver_port
  mode tcp
  option tcplog
  default_backend kube-apiserver

listen stats
  bind *:8888
  mode http
  stats enable
  stats uri /stats
  stats refresh 5s
  stats show-node
  stats auth kcai:kcai1212
  stats realm Haproxy\ Statistics
  log 127.0.0.1 local3 err

backend kube-apiserver
  mode tcp
  balance roundrobin
EOF
  for node in "${master_nodes[@]}"; do
    name=$(echo "$node" | awk '{print $1}')
    ip_port=$(echo "$node" | awk '{print $2}')
    echo "  server $name $ip_port check" >> haproxy.cfg
  done
  # 以docker容器方式运行haproxy
  echo "以docker容器方式运行haproxy"
  docker run -d --name haproxy \
  --net=host \
  --restart=always \
  -v "$(pwd)"/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro \
  haproxytech/haproxy-ubuntu:2.8
fi

# 配置keepalived
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  echo "配置keepalived"
  cat <<EOF > keepalived.conf
global_defs {
  router_id VI_1_$node_name
}

vrrp_script check_haproxy {
  script /usr/bin/check_haproxy.sh
  interval 2
  weight -30
}

vrrp_instance VI_1 {
  state $keepalived_state
  interface $net_card_name
  virtual_router_id 51
  priority 100
  advert_int 1

  authentication {
    auth_type PASS
    auth_pass kcai1212
  }

  virtual_ipaddress {
    $keepalived_vip/24 dev $net_card_name
  }

  track_script {
    check_haproxy
  }
}
EOF
  # 配置check_haproxy.sh
  echo "配置check_haproxy.sh"
  cat <<EOF > check_haproxy.sh
#!/bin/bash
count=`netstat -apn | grep 9443 | grep haproxy | wc -l`
if [ \$count -gt 0 ]; then
  exit 0
else
  exit 1
fi
EOF
  chmod +x check_haproxy.sh
  # 以docker容器方式运行keepalived
  echo "以docker容器方式运行keepalived"
  docker run -d --name keepalived \
  --net=host \
  --cap-add=NET_ADMIN \
  --cap-add=NET_BROADCAST \
  --cap-add=NET_RAW \
  --restart=always \
  -v "$(pwd)"/keepalived.conf:/container/service/keepalived/assets/keepalived.conf \
  -v "$(pwd)"/check_haproxy.sh:/usr/bin/check_haproxy.sh \
  osixia/keepalived:2.0.20 --copy-service
fi

# 安装kubernetes相关组件
echo "安装kubernetes相关组件"
curl https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
EOF
apt-get update
apt-get install -y kubelet=1.28.1-00 kubeadm=1.28.1-00
if [[ "$init_type" == "1" || "$init_type" == "2" ]]; then
  apt-get install -y kubectl=1.28.1-00
fi

# 创建集群
if [[ "$init_type" == "1" ]]; then
  # 设置kubernetes参数
  echo "设置kubernetes参数"
  cat <<EOF > init-defaults.k8s.yaml
kind: InitConfiguration
apiVersion: kubeadm.k8s.io/v1beta3
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  token: abcdef.0123456789abcdef
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
localAPIEndpoint:
  advertiseAddress: $this_ip
  bindPort: 6443
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
  imagePullPolicy: IfNotPresent
  name: $node_name
  taints: []
---
kind: ClusterConfiguration
apiVersion: kubeadm.k8s.io/v1beta3
certificatesDir: /etc/kubernetes/pki
clusterName: kubernetes
controllerManager: {}
dns: {}
etcd:
  local:
    dataDir: /var/lib/etcd
imageRepository: registry.cn-hangzhou.aliyuncs.com/google_containers
kubernetesVersion: 1.28.0
networking:
  dnsDomain: cluster.local
  serviceSubnet: 10.96.0.0/12
scheduler: {}
controlPlaneEndpoint: $apiserver_domain:$apiserver_port
apiServer:
  timeoutForControlPlane: 4m0s
  certSANs:
  - kubernetes
  - kubernetes.default
  - kubernetes.default.svc
  - kubernetes.default.svc.cluster.local
  - localhost
  - 127.0.0.1
  - protodesign.cn
  - www.protodesign.cn
  - api.protodesign.cn
  - test.protodesign.cn
  - $apiserver_domain
  - $apiserver_ip
EOF
  for node in "${master_nodes[@]}"; do
    name=$(echo "$node" | awk '{print $1}')
    ip_port=$(echo "$node" | awk '{print $2}')
    ip=$(echo "$ip_port" | cut -d: -f1)
    echo "  - $name" >> init-defaults.k8s.yaml
    echo "  - $ip" >> init-defaults.k8s.yaml
  done
  # 手动修改参数：name、advertiseAddress等
  read -r -p "是否需要手动修改kubeadm init config参数（y/n）" init_type
  if [[ "$init_type" == "" || "$init_type" == "y" ]]; then
      vim init-defaults.k8s.yaml
  fi

  # 初始化kubernetes集群
  echo "初始化kubernetes集群"
  kubeadm init --config init-defaults.k8s.yaml --upload-certs | tee kubeadm-init.log
  mkdir -p $HOME/.kube
  cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  chown "$(id -u):$(id -g)" $HOME/.kube/config

  # 下载配置文件 calico
  echo "下载配置文件 calico"
  export http_proxy=$proxy_address
  export https_proxy=$proxy_address
  export HTTP_PROXY=$proxy_address
  export HTTPS_PROXY=$proxy_address
  curl https://docs.projectcalico.org/manifests/calico.yaml -LO
  export HTTP_PROXY=
  export HTTPS_PROXY=
  export http_proxy=
  export https_proxy=

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
  cat <<EOF > join-defaults.k8s.yaml
kind: JoinConfiguration
apiVersion: kubeadm.k8s.io/v1beta3
discovery:
  bootstrapToken:
    token: $join_token
    apiServerEndpoint: $apiserver_domain:$apiserver_port
    unsafeSkipCAVerification: true
  timeout: 5m0s
  tlsBootstrapToken: $join_token
nodeRegistration:
  name: $node_name
  criSocket: unix:///var/run/containerd/containerd.sock
  imagePullPolicy: IfNotPresent
  taints: []
caCertPath: "/etc/kubernetes/pki/ca.crt"
EOF
  if [ "$init_type" == "2" ]; then # 加入master节点
  cat <<EOF >> join-defaults.k8s.yaml
controlPlane:
  localAPIEndpoint:
    advertiseAddress: $this_ip
    bindPort: 6443
  certificateKey: $certificate_key
EOF
  fi
  # 加入集群
  echo "加入集群"
  kubeadm join --config join-defaults.k8s.yaml
  # 设置kubectl
  if [[ "$init_type" == "2" ]]; then
    mkdir -p $HOME/.kube
    cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    chown "$(id -u):$(id -g)" $HOME/.kube/config
    # 清除master节点污点
    echo "清除master节点污点"
    kubectl taint nodes --all node-role.kubernetes.io/master-
  fi
fi
