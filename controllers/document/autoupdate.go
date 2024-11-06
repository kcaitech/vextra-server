package document

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
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
// const minUpdateTimeInterval time.Duration = time.Second * 60 * 10

type DocumentVersioningInfo struct {
	DocId          int64     `json:"docId"`
	LastUpdateTime time.Time `json:"lastUpdateTime"`
}

// 并不是同一个文档的都在一个服务实例里，也就个人编辑有点用
var documentVersioningInfoMap = my_map.NewSyncMap[int64, DocumentVersioningInfo]()

func getDocumentLastUpdateTimeFromRedis(documentId int64) time.Time {
	if lastUpdateTime, err := redis.Client.Get(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+str.IntToString(documentId)+"]").Int64(); err == nil && lastUpdateTime > 0 {
		return time.UnixMilli(lastUpdateTime)
	}
	return time.UnixMilli(0)
}

// body: { documentInfo: DocumentInfo, lastCmdId: string, documentData: ExFromJson, documentText: string, mediasSize: number, pageImageBase64List: string[] }

type DocumentInfo struct {
	DocumentId string `json:"id"` // 可能是int
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
	// DocumentInfo DocumentInfo `json:"documentInfo"`
	LastCmdId    string     `json:"lastCmdId"`
	DocumentData ExFromJson `json:"documentData"`
	DocumentText string     `json:"documentText"`
	MediasSize   uint64     `json:"mediasSize"`
	PageSvgs     []string   `json:"pageSvgs"`
}

func _svgToPng(svgContent string) ([]byte, error) {
	// 将SVG内容转换为字节切片
	svgBuffer := []byte(svgContent)

	// 创建FormData
	formData := new(bytes.Buffer)
	writer := multipart.NewWriter(formData)

	// 添加SVG文件到表单
	part, err := writer.CreateFormFile("svg", "image.svg")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, bytes.NewReader(svgBuffer))
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// 设置请求头
	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	url := config.Config.Svg2Png.Url

	// 创建请求
	req, err := http.NewRequest("POST", url, formData)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 设置响应类型为 arraybuffer
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SVG to PNG conversion failed with status: %d, message: %s", resp.StatusCode, body)
		return nil, fmt.Errorf("svgToPng错误")
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func svg2png(svgs []string) *[][]byte {

	// 创建一个 WaitGroup 并设置初始计数器值
	var wg sync.WaitGroup

	count := len(svgs)
	wg.Add(count) // 假设我们要发起5个请求

	pngs := make([][]byte, count)

	// 创建一个有界通道，用于限制并发请求的数量
	const maxConcurrentRequests = 5
	requestCh := make(chan struct{}, maxConcurrentRequests)

	// 循环发起请求
	for i, svg := range svgs {
		go func(i int, svg string) {
			defer wg.Done() // 每个 goroutine 结束时调用 Done()

			// 获取一个并发请求的许可
			requestCh <- struct{}{}

			// 在请求结束后释放许可
			defer func() {
				<-requestCh
			}()

			png, err := _svgToPng(svg)
			if err != nil {
				log.Println("svg2png fail", err)
			} else {
				pngs[i] = png
			}
		}(i, svg)
	}

	// 等待所有请求完成
	wg.Wait()
	// fmt.Println("All requests have been processed.")
	return &pngs
}

func AutoUpdate(documentId int64) {
	info, ok := documentVersioningInfoMap.Get(documentId)
	if !ok {
		info = DocumentVersioningInfo{
			DocId:          documentId,
			LastUpdateTime: time.UnixMilli(0),
		}
		documentVersioningInfoMap.Set(documentId, info)
	}
	minUpdateTimeInterval := time.Second * time.Duration(config.Config.VersionServer.MinUpdateInterval)
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
	defer func() {
		if _, err := documentVersioningMutex.Unlock(); err != nil {
			log.Println(documentId, "释放锁失败 documentVersioningMutex.Unlock", err)
		}
	}()
	// 从redis获取LastUpdateTime，更新到本地缓存
	lastUpdateTimeFromRedis := getDocumentLastUpdateTimeFromRedis(documentId)
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
	var generateApiUrl = config.Config.VersionServer.Url

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

	log.Println("auto update document, start svg2png")
	// svg2png
	pagePngs := svg2png(version.PageSvgs)

	log.Println("auto update document, start upload data", documentId)
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
		MediaNames:    version.DocumentData.MediaNames,
		MediasSize:    version.MediasSize,
		DocumentText:  version.DocumentText,
		PageImageList: pagePngs,
	}
	UploadDocumentData(&header, &data, nil, &response)

	if response.Status != ResponseStatusSuccess {
		log.Println("UploadDocumentData fail")
		return
	}

	// 更新redis
	if _, err := redis.Client.Set(context.Background(), "Document Versioning LastUpdateTime[DocumentId:"+documentIdStr+"]", time.Now().UnixMilli(), time.Hour*1).Result(); err != nil {
		log.Println("redis.Client.Set err", err)
	} else {
		log.Println("auto update successed")
	}
}
