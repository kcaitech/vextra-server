#!/bin/bash

kubectl apply -f local-hostpath-pvc.yaml
kubectl apply -f local-hostpath-pod.yaml
echo "sleep 10 ..."
sleep 10
echo "kubectl get pod hello-local-hostpath-pod"
kubectl get pod hello-local-hostpath-pod
echo "kubectl exec hello-local-hostpath-pod -- cat /mnt/store/greet.txt"
kubectl exec hello-local-hostpath-pod -- cat /mnt/store/greet.txt
echo "kubectl describe pod hello-local-hostpath-pod"
kubectl describe pod hello-local-hostpath-pod
echo "kubectl get pvc local-hostpath-pvc"
kubectl get pvc local-hostpath-pvc
echo "kubectl get pv pvc-... -o yaml"
echo "可以去挂在节点的对应目录下查看（ls /var/local-hostpath）"
echo "kubectl delete -f local-hostpath-pod.yaml"
echo "kubectl delete -f local-hostpath-pvc.yaml"
