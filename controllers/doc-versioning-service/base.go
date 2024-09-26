package doc_versioning_service

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-redsync/redsync/v4"
	"io"
	"log"
	"net/http"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/utils/my_map"
	"kcaitech.com/kcserver/utils/str"
	"time"
	// "kcaitech.com/kcserver/common"
	config "kcaitech.com/kcserver/controllers"
)

// 最短更新时间间隔（秒）
const minUpdateTimeInterval time.Duration = time.Second * 60 * 10

type DocumentVersioningInfo struct {
	DocId          int64     `json:"docId"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

var documentVersioningInfoMap = my_map.NewSyncMap[int64, DocumentVersioningInfo]()

func getDocumentLastUpdateTimeFromRedis(documentId int64) time.Time {
	if lastUpdateTime, err := redis.Client.Get(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+str.IntToString(documentId)+"]").Int64(); err == nil && lastUpdateTime > 0 {
		return time.UnixMilli(lastUpdateTime)
	}
	return time.UnixMilli(0)
}

var docVersioningServiceUrl = config.Config.DocumentVersionServer.Host
var generateApiUrl = "http://" + docVersioningServiceUrl + "/generate"

func Trigger(documentId int64) {
	info, ok := documentVersioningInfoMap.Get(documentId)
	if !ok {
		info = DocumentVersioningInfo{
			DocId:          documentId,
			LastUpdateTime: time.UnixMilli(0),
		}
		documentVersioningInfoMap.Set(documentId, info)
	}
	// 时间未到
	if time.Since(info.LastUpdateTime) < minUpdateTimeInterval {
		return
	}
	// 上锁
	documentIdStr := str.IntToString(documentId)
	documentVersioningMutex := redis.RedSync.NewMutex("Document Versioning Mutex[DocumentId:"+documentIdStr+"]", redsync.WithExpiry(time.Second*10))
	if err := documentVersioningMutex.TryLock(); err != nil {
		info.LastUpdateTime = time.Now()
		return
	}
	// 从redis获取LastUpdateTime，更新到本地缓存
	lastUpdateTimeFromRedis := getDocumentLastUpdateTimeFromRedis(documentId)
	if !lastUpdateTimeFromRedis.IsZero() {
		info.LastUpdateTime = lastUpdateTimeFromRedis
	}
	// 再检测一遍，时间未到
	if time.Since(info.LastUpdateTime) < minUpdateTimeInterval {
		return
	}
	defer func() {
		if _, err := documentVersioningMutex.Unlock(); err != nil {
			log.Println(documentId, "释放锁失败 documentVersioningMutex.Unlock", err)
		}
	}()
	// 开始更新版本
	defer func() {
		info.LastUpdateTime = time.Now()
	}()
	requestData := map[string]any{
		"documentId": documentIdStr,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return
	}
	// 构建请求
	req, err := http.NewRequest("POST", generateApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(generateApiUrl, "http.NewRequest err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(generateApiUrl, "client.Do err", err)
		return
	}
	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(generateApiUrl, "io.ReadAll err", err)
		return
	}
	if resp.StatusCode != 200 || string(body) != "success" {
		log.Println(generateApiUrl, "请求失败", resp.StatusCode, string(body))
	}
	// 更新redis
	if _, err := redis.Client.Set(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+documentIdStr+"]", time.Now().UnixMilli(), time.Hour*1).Result(); err != nil {
		log.Println("redis.Client.Set err", err)
	}
}
