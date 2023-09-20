#!/bin/bash

ipaddr=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}' | awk 'NR==1{print}')

echo "current ip address is ${ipaddr}"

echo "create file /root/kuboard-sa.yaml"

echo

cat > /root/kuboard-sa.yaml << EOF
---
kind: Namespace
apiVersion: v1
metadata:
  name: kuboard

---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: kuboard-admin
  namespace: kuboard

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kuboard-admin-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: kuboard-admin
    namespace: kuboard

---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: kuboard-viewer
  namespace: kuboard

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kuboard-viewer-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
  - kind: ServiceAccount
    name: kuboard-viewer
    namespace: kuboard

EOF

echo "kubectl apply -f /root/kuboard-sa.yaml"

kubectl apply -f /root/kuboard-sa.yaml

echo
echo "create file /etc/kubernetes/manifests/kuboard.yaml"
echo

cat > /etc/kubernetes/manifests/kuboard.yaml << EOF
---
apiVersion: v1
kind: Pod
metadata:
  annotations: {}
  labels:
    k8s.kuboard.cn/name: kuboard-v3
  name: kuboard-v3
  namespace: kuboard
spec:
  containers:
    - env:
        - name: KUBOARD_ENDPOINT
          value: "http://${ipaddr}:80"
        - name: KUBOARD_AGENT_SERVER_TCP_PORT
          value: "10081"
      image: 'eipwork/kuboard:v3'
      imagePullPolicy: IfNotPresent
      livenessProbe:
        failureThreshold: 3
        httpGet:
          path: /kuboard-resources/version.json
          port: 80
          scheme: HTTP
        initialDelaySeconds: 30
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
      name: kuboard
      ports:
        - containerPort: 80
          hostPort: 30001
          name: web
          protocol: TCP
        - containerPort: 10081
          name: peer
          protocol: TCP
          hostPort: 10081
        - containerPort: 10081
          name: peer-u
          protocol: UDP
          hostPort: 10081
      readinessProbe:
        failureThreshold: 3
        httpGet:
          path: /kuboard-resources/version.json
          port: 80
          scheme: HTTP
        initialDelaySeconds: 30
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
      volumeMounts:
        - mountPath: /data
          name: data
        - mountPath: /init-etcd-scripts/import-cluster-once.yaml
          name: import-cluster-yaml
  volumes:
    - hostPath:
        path: "/usr/share/kuboard"
      name: data
    - hostPath:
        path: "/usr/share/kuboard/import-cluster-once.yaml"
      name: import-cluster-yaml
  dnsPolicy: ClusterFirst
  restartPolicy: Always
  tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
EOF

echo "restart kubelet"

systemctl restart kubelet

host_name=$(hostname)

echo
echo "\033[34m检查状态\033[0m 待 kuboard-v3-${host_name} 的容器组变为 Running 状态后，则安装成功，可以通过 http://${ipaddr}:30001 访问 kuboard 界面"
echo "初始账号密码：admin Kuboard123"
echo

kubectl get pods -n kuboard -o wide