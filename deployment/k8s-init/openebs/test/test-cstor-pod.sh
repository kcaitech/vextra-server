#!/bin/bash

# 测试cStor

kubectl apply -f cstor-pod.yaml
echo "cStor池实例（cspi/CStorPoolInstance）状态："
echo "kubectl -n openebs get cspi"
echo "cStor卷（cv/CStorVolume）状态："
echo "kubectl -n openebs get cv"
echo "cStor卷副本（cvr/CStorVolumeReplica）状态："
echo "kubectl -n openebs get cvr"
echo "清除测试环境："
echo "kubectl delete -f cstor-pod.yaml"
