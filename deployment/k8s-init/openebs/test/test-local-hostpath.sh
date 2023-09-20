#!/bin/bash

kubectl apply -f local-hostpath-pvc.yaml
kubectl apply -f local-hostpath-pod.yaml
echo "sleep 10 ..."
sleep 10
echo "kubectl get pod test-local-hostpath"
kubectl get pod test-local-hostpath
echo "kubectl exec test-local-hostpath -- cat /mnt/store/greet.txt"
kubectl exec test-local-hostpath -- cat /mnt/store/greet.txt
echo "kubectl describe pod test-local-hostpath"
kubectl describe pod test-local-hostpath
echo "kubectl get pvc local-hostpath-pvc"
kubectl get pvc local-hostpath-pvc
echo "kubectl get pv pvc-... -o yaml"
echo "可以去挂载节点的对应目录下查看（ls /var/local-hostpath）"
echo "kubectl delete -f local-hostpath-pod.yaml"
echo "kubectl delete -f local-hostpath-pvc.yaml"
