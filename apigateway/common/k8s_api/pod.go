package k8s_api

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"protodesign.cn/kcserver/utils/sliceutil"
	"time"
)

var clientSet *kubernetes.Clientset

func Init() error {
	var err error
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.New("获取集群配置失败: " + err.Error())
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.New("创建Kubernetes客户端失败: " + err.Error())
	}
	return nil
}

var docOpNS = "kc"
var docOpPodLabel = "app=docop-server-czf"

func GetDocOpPods() []string {
	pods, err := clientSet.CoreV1().Pods(docOpNS).List(context.TODO(), metav1.ListOptions{
		LabelSelector: docOpPodLabel,
	})
	if err != nil {
		log.Println("获取docop pod列表失败", err)
		return nil
	}
	return sliceutil.MapT(func(pod corev1.Pod) string {
		return pod.Name
	}, pods.Items...)
}

var docOpPodsCache []string
var docOpPodsCacheTime int64

// 毫秒
var docOpPodsCacheTimeOut int64 = 1000 * 3

func GetDocOpPodsCache() []string {
	now := time.Now().UnixNano() / 1000000
	if now-docOpPodsCacheTime > docOpPodsCacheTimeOut {
		docOpPodsCache = GetDocOpPods()
		docOpPodsCacheTime = now
	}
	return docOpPodsCache
}
