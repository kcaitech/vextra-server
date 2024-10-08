package document

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-redsync/redsync/v4"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/utils/my_map"
	"kcaitech.com/kcserver/utils/str"

	// "kcaitech.com/kcserver/common"
	config "kcaitech.com/kcserver/controllers"
	// document "kcaitech.com/kcserver/controllers/document"
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

// body: { documentInfo: DocumentInfo, lastCmdId: string, documentData: ExFromJson, documentText: string, mediasSize: number, pageImageBase64List: string[] }

type DocumentInfo struct {
	DocumentId string `json:"id"`
	Path       string `json:"path"`
	VersionId  string `json:"version_id"`
	LastCmdId  string `json:"last_cmd_id"`
}

type ExFromJson struct {
	DocumentMeta Data            `json:"document_meta"`
	Pages        json.RawMessage `json:"pages"`
	MediaNames   []string        `json:"media_names"`

	// FreeSymbols         json.RawMessage `json:"freesymbols"` // 这个在DocumentMeta里
	// MediasSize          uint64          `json:"medias_size"`
	// DocumentText        string          `json:"document_text"`
	// PageImageBase64List []string        `json:"page_image_base64_list"`
}

type VersionResp struct {
	DocumentInfo        DocumentInfo `json:"documentInfo"`
	LastCmdId           string       `json:"lastCmdId"`
	DocumentData        ExFromJson   `json:"documentData"`
	DocumentText        string       `json:"documentText"`
	MediasSize          uint64       `json:"mediasSize"`
	PageImageBase64List []string     `json:"pageImageBase64List"`
}

func AutoSave(documentId int64) {
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
	// requestData := map[string]any{
	// 	"documentId": documentIdStr,
	// }
	// jsonData, err := json.Marshal(requestData)
	// if err != nil {
	// 	return
	// }
	// 构建请求
	resp, err := http.Get(generateApiUrl + "?documentId=" + documentIdStr)
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
	if resp.StatusCode != 200 || string(body) != "success" {
		log.Println(generateApiUrl, "请求失败", resp.StatusCode, string(body))
		return
	}

	version := VersionResp{}
	err = json.Unmarshal(body, &version)
	if err != nil {
		log.Println(generateApiUrl, "resp", err)
		return
	}

	// upload document data
	header := Header{
		DocumentId: documentIdStr,
		LastCmdId:  version.LastCmdId,
	}
	response := Response{}
	data := UploadData{
		DocumentMeta: Data(version.DocumentData.DocumentMeta),
		Pages:        version.DocumentData.Pages,
		// FreeSymbols        : version.DocumentData.DocumentMeta.
		MediaNames:          version.DocumentData.MediaNames,
		MediasSize:          version.MediasSize,
		DocumentText:        version.DocumentText,
		PageImageBase64List: version.PageImageBase64List,
	}
	UploadDocumentData(&header, &data, nil, &response)

	if response.Status != ResponseStatusSuccess {
		log.Println("UploadDocumentData fail")
		return
	}

	// 更新redis
	if _, err := redis.Client.Set(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+documentIdStr+"]", time.Now().UnixMilli(), time.Hour*1).Result(); err != nil {
		log.Println("redis.Client.Set err", err)
	}
}
