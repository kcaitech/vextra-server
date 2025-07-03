package common

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redsync/redsync/v4"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/redis"
	"kcaitech.com/kcserver/utils/my_map"

	// "kcaitech.com/kcserver/common"
	config "kcaitech.com/kcserver/config"
	"kcaitech.com/kcserver/services"
	// document "kcaitech.com/kcserver/handlers/document"
)

// 最短更新时间间隔（秒）
// const minUpdateTimeInterval time.Duration = time.Second * 60 * 10

type DocumentVersioningInfo struct {
	DocId          string    `json:"docId"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// 并不是同一个文档的都在一个服务实例里，也就个人编辑有点用
var documentVersioningInfoMap = my_map.NewSyncMap[string, DocumentVersioningInfo]()

func getDocumentLastUpdateTimeFromRedis(documentId string, redis *redis.RedisDB) time.Time {
	if lastUpdateTime, err := redis.Client.Get(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+(documentId)+"]").Int64(); err == nil && lastUpdateTime > 0 {
		return time.UnixMilli(lastUpdateTime)
	}
	return time.UnixMilli(0)
}

type DocumentInfo struct {
	DocumentId string `json:"id"` // 可能是int
	Path       string `json:"path"`
	VersionId  string `json:"version_id"`
	LastCmdId  uint   `json:"last_cmd_id"`
}

type ExFromJson struct {
	DocumentMeta Data            `json:"document_meta"`
	Pages        json.RawMessage `json:"pages"`
	MediaNames   []string        `json:"media_names"`
}

type VersionResp struct {
	LastCmdVerId string     `json:"lastCmdVerId"`
	DocumentData ExFromJson `json:"documentData"`
	DocumentText string     `json:"documentText"`
	MediasSize   uint64     `json:"mediasSize"`
	PagePngs     []string   `json:"pages_png_generated"`
	TmpPngDir    string     `json:"tmp_png_dir"`
}

func AutoUpdate(documentId string, config *config.Configuration) {
	info, ok := documentVersioningInfoMap.Get(documentId)
	if !ok {
		info = DocumentVersioningInfo{
			DocId:          documentId,
			LastUpdateTime: time.UnixMilli(0),
		}
		documentVersioningInfoMap.Set(documentId, info)
	}
	minUpdateTimeInterval := time.Second * time.Duration(config.VersionServer.MinUpdateInterval)
	// 时间未到
	if time.Since(info.LastUpdateTime) < minUpdateTimeInterval {
		return
	}
	// 上锁
	// documentIdStr := str.IntToString(documentId)
	redis := services.GetRedisDB()
	documentVersioningMutex := redis.RedSync.NewMutex("Document Versioning Mutex[DocumentId:"+documentId+"]", redsync.WithExpiry(time.Second*10))
	if err := documentVersioningMutex.TryLock(); err != nil {
		info.LastUpdateTime = time.Now()
		return
	}
	defer func() {
		if _, err := documentVersioningMutex.Unlock(); err != nil {
			log.Println(documentId, "释放锁失败 documentVersioningMutex.Unlock", err)
		}
	}()
	// 从redis获取LastUpdateTime，更新到本地缓存
	lastUpdateTimeFromRedis := getDocumentLastUpdateTimeFromRedis(documentId, redis)
	if !lastUpdateTimeFromRedis.IsZero() {
		info.LastUpdateTime = lastUpdateTimeFromRedis
	}
	// 再检测一遍，时间未到
	if time.Since(info.LastUpdateTime) < minUpdateTimeInterval {
		return
	}
	// 开始更新版本
	defer func() {
		info.LastUpdateTime = time.Now()
	}()

	log.Println("auto update document:", documentId)
	var generateApiUrl = config.VersionServer.Url

	documentInfo, err := GetDocumentBasicInfoById(documentId)

	if err != nil {
		log.Println("获取文档信息失败: " + err.Error())
		return
	}

	cmdService := services.GetCmdService()
	lastCmdId := documentInfo.LastCmdId + 1
	cmdItemList, err := cmdService.GetCmdItemsFromStart(documentId, lastCmdId)

	if err != nil {
		log.Println("获取命令列表失败: " + err.Error())
		return
	}

	if len(cmdItemList) == 0 {
		log.Println("没有命令需要更新版本")
		return
	}

	if len(cmdItemList) < config.VersionServer.MinCmdCount {
		log.Println("命令数量小于", config.VersionServer.MinCmdCount, "不更新版本")
		return
	}
	// 构建请求
	reqBody := map[string]interface{}{
		"documentInfo": documentInfo,
		"cmdItemList":  cmdItemList,
		// "force":        false,
	}

	tmpPngDir := config.SafeReview.TmpPngDir + "/" + documentId
	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil { // 需要审查才生成png图片
		reqBody["gen_pages_png"] = map[string]interface{}{
			"tmp_dir": tmpPngDir,
		}
		os.MkdirAll(tmpPngDir, 0755)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Println("Failed to marshal request body:", err)
		return
	}

	resp, err := http.Post(generateApiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(generateApiUrl, "http.NewRequest err", err)
		return
	}

	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(generateApiUrl, "io.ReadAll err", err)
		return
	}
	if resp.StatusCode != 200 {
		log.Println(generateApiUrl, "请求失败", resp.StatusCode, string(body))
		return
	}
	version := VersionResp{}
	err = json.Unmarshal(body, &version)
	if err != nil {
		log.Println(generateApiUrl, "resp", err)
		return
	}

	version.TmpPngDir = tmpPngDir

	log.Println("auto update document, start upload data", documentId)
	// upload document data
	// header := Header{
	// 	DocumentId:   documentId,
	// 	LastCmdVerId: version.LastCmdVerId,
	// }
	response := Response{}
	UpdateDocumentData(documentId, version.LastCmdVerId, &version, nil, &response)

	if response.Code != http.StatusOK {
		log.Println("UploadDocumentData fail")
		return
	}

	if publishData, err := json.Marshal(&models.DocumentVersionWSData{
		DocumentId:       documentId,
		VersionId:        documentInfo.VersionId,
		VersionStartWith: lastCmdId,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Version[DocumentId:"+(documentId)+"]", publishData)
	}

	// 更新redis
	if _, err := redis.Client.Set(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+documentId+"]", time.Now().UnixMilli(), time.Hour*1).Result(); err != nil {
		log.Println("redis.Client.Set err", err)
	} else {
		log.Println("auto update successed")
	}
}
