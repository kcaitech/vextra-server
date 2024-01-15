package docop

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
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
	res := redis.Client.Get(context.Background(), "DocopServer:pod[document:"+documentId+"]")
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
	if len(podInfoList) == 0 {
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

const docopServiceUrl = "docop-server-headless.kc.svc.cluster.local"

func GetDocumentUrl(documentId string) string {
	pod := GetPodByDocumentId(documentId)
	if pod != "" {
		return "ws://" + pod + "." + docopServiceUrl + ":10011"
	}

	pod = GetPodByMinDocument()
	if pod == "" {
		return ""
	}

	managerAddUrl := "http://" + pod + "." + docopServiceUrl + ":10010" + "/add"

	requestData := map[string]any{
		"documentId": documentId,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return ""
	}

	req, err := http.NewRequest("POST", managerAddUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(managerAddUrl, "http.NewRequest err", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(managerAddUrl, "client.Do err", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(managerAddUrl, "io.ReadAll err", err)
		return ""
	}

	if resp.StatusCode != 200 || string(body) != "success" {
		log.Println(managerAddUrl, "请求失败", resp.StatusCode, string(body))
		return ""
	}

	return "ws://" + pod + "." + docopServiceUrl + ":10011"
}

const getDocumentUrlRetryCount = 3
const getDocumentUrlRetryInterval = 1

func GetDocumentUrlRetry(documentId string) string {
	for i := 0; i < getDocumentUrlRetryCount; i++ {
		url := GetDocumentUrl(documentId)
		if url != "" {
			return url
		}
		if i < getDocumentUrlRetryCount-1 {
			time.Sleep(time.Second * getDocumentUrlRetryInterval)
		}
	}
	return ""
}
