#!/bin/bash

set -e

echo "安装docker"

#for pkg in docker.io docker-doc docker-compose podman-docker containerd runc; do apt remove -y $pkg; done

apt install -y ca-certificates curl gnupg

rm -f /etc/apt/keyrings/docker.gpg
install -m 0755 -d /etc/apt/keyrings

# curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

chmod a+r /etc/apt/keyrings/docker.gpg

#echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
#$(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://mirrors.aliyun.com/docker-ce/linux/debian \
$(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin docker-compose
# 设置docker参数
echo "设置docker参数"
cp etc_docker_daemon.json /etc/docker/daemon.json
systemctl enable --now docker
# 重启docker服务
echo "重启docker服务"
systemctl restart docker
# 验证docker服务是否正常
#echo "验证docker服务是否正常"
#docker images
#docker pull busybox
#docker run --rm busybox uname -m
