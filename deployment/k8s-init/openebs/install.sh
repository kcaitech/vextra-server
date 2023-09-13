#!/bin/bash

# 安装OpenEBS

# 添加仓库
helm repo add openebs https://openebs.github.io/charts
helm repo update
# 安装
helm install openebs openebs/openebs -n openebs --create-namespace -f values.yaml

# 创建storageclass
kubectl apply -f local-hostpath-sc.yaml

# 创建cStor池
echo "请等待openebs服务创建完毕再执行以下操作："

echo "块设备信息："
echo "kubectl -n openebs get bd"
echo "请根据块设备信息修改cstor-pool-cluster.yaml文件，修改后创建cStor池："
echo "kubectl apply -f cstor-pool-cluster.yaml"
echo "查看cStor池状态："
echo "kubectl -n openebs get cspc"
echo "kubectl -n openebs get cspi"
echo "创建cStor存储类："
echo "kubectl apply -f cstor-sc.yaml"
echo "查看cStor存储类状态："
echo "kubectl get sc cstor-sc"
