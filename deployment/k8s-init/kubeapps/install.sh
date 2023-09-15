#!/bin/bash

# 安装kubeapps

set -e

# 添加bitnami helm仓库
helm repo add bitnami https://charts.bitnami.com/bitnami
# 安装
helm install kubeapps bitnami/kubeapps -n kubeapps --create-namespace
kubectl apply -f nodeport-services.yaml
# 创建访问密钥
kubectl create --namespace default serviceaccount kubeapps-operator
kubectl create clusterrolebinding kubeapps-operator --clusterrole=cluster-admin --serviceaccount=default:kubeapps-operator
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: kubeapps-operator-token
  namespace: default
  annotations:
    kubernetes.io/service-account.name: kubeapps-operator
type: kubernetes.io/service-account-token
EOF
# 输出token
echo "端口号：30001，登录token："
kubectl get --namespace default secret kubeapps-operator-token -o go-template='{{.data.token | base64decode}}'

# kubectl port-forward --namespace kubeapps service/kubeapps 8080:80
# 运行上述命令后打开 http://127.0.0.1:8080
