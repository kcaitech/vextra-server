package docop

import (
	"context"
	"protodesign.cn/kcserver/apigateway/common/k8s_api"
	"protodesign.cn/kcserver/common/redis"
	"protodesign.cn/kcserver/utils/set"
	"protodesign.cn/kcserver/utils/sliceutil"
	"time"
)

func delPodFromRedis(podName string) {
	redis.Client.SRem(context.Background(), "DocopServer:podSet", podName)
	documentList := redis.Client.SMembers(context.Background(), "DocopServer:documentSet[pod:"+podName+"]").Val()
	redis.Client.Del(context.Background(), "DocopServer:documentSet[pod:"+podName+"]")
	redis.Client.SRem(context.Background(), "DocopServer:documentSet", sliceutil.ToAny(documentList)...)
	podKeyList := sliceutil.MapT(func(documentId string) string {
		return "DocopServer:pod[document:" + documentId + "]"
	}, documentList...)
	redis.Client.Del(context.Background(), podKeyList...)
}

func GetPodByDocumentId(documentId string) string {
	res := redis.Client.Get(context.Background(), "DocopServer:Pod[document:"+documentId+"]")
	if res.Err() != nil {
		return ""
	}
	podName := res.Val()

	if podName == "" {
		redis.Client.Del(context.Background(), "DocopServer:pod[document:"+documentId+"]")
		redis.Client.SRem(context.Background(), "DocopServer:documentSet", documentId)
		return ""
	}
	if !k8s_api.ExistsDocOpPod(podName) {
		redis.Client.Del(context.Background(), "DocopServer:pod[document:"+documentId+"]")
		redis.Client.SRem(context.Background(), "DocopServer:documentSet", documentId)
		delPodFromRedis(podName)
		return ""
	}

	return podName
}

func GetPods() []string {
	podList := redis.Client.SMembers(context.Background(), "DocopServer:podSet").Val()
	pods := set.NewSet(podList...)
	k8sPods := set.NewSet(k8s_api.GetDocOpPodsDefault()...)
	diffPods := pods.Difference(k8sPods)
	if diffPods.Size() == 0 {
		return podList
	}
	k8sPods = set.NewSet(k8s_api.GetDocOpPods()...)
	diffPods = pods.Difference(k8sPods)
	if diffPods.Size() == 0 {
		return podList
	}
	for _, pod := range diffPods.Items() {
		delPodFromRedis(pod)
	}
	return nil
}

type PodInfo struct {
	PodName     string
	DocumentIds []string
	UpdateTime  int64
}

var podInfoMap = make(map[string]*PodInfo)

func GetPodInfo(podName string) *PodInfo {
	now := time.Now().UnixNano() / 1000000
	info, ok := podInfoMap[podName]
	if ok && now-info.UpdateTime < 1000*3 {
		return info
	}
	res := redis.Client.SIsMember(context.Background(), "DocopServer:podSet", podName)
	if res.Err() != nil || !res.Val() {
		delete(podInfoMap, podName)
		return nil
	}
	documentList := redis.Client.SMembers(context.Background(), "DocopServer:documentSet[pod:"+podName+"]").Val()
	if !ok {
		info = &PodInfo{
			PodName:     podName,
			DocumentIds: documentList,
			UpdateTime:  now,
		}
		podInfoMap[podName] = info
	} else {
		info.DocumentIds = documentList
		info.UpdateTime = now
	}
	return info
}

func GetPodsInfo() []*PodInfo {
	podList := GetPods()
	if podList == nil {
		return nil
	}
	var podInfoList []*PodInfo
	for _, podName := range podList {
		podInfo := GetPodInfo(podName)
		if podInfo != nil {
			podInfoList = append(podInfoList, podInfo)
		}
	}
	return podInfoList
}

func GetPodByMinDocument() string {
	podInfoList := GetPodsInfo()
	if podInfoList == nil {
		return ""
	}
	minDocumentPod := podInfoList[0]
	for _, podInfo := range podInfoList {
		if len(podInfo.DocumentIds) < len(minDocumentPod.DocumentIds) {
			minDocumentPod = podInfo
		}
	}
	return minDocumentPod.PodName
}
