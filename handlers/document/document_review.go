package document

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/str"
)

func _svgToPng(svgContent string, svg2pngUrl string) ([]byte, error) {
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

	// 创建请求
	req, err := http.NewRequest("POST", svg2pngUrl, formData)
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

func svg2png(uploadData *UploadData, svg2pngUrl string) *[][]byte {
	count := len(uploadData.PageSvgs)
	if count == 0 {
		return nil
	}

	// 创建一个 WaitGroup 并设置初始计数器值
	var wg sync.WaitGroup

	wg.Add(count) // 假设我们要发起5个请求

	pngs := make([][]byte, count)

	// 创建一个有界通道，用于限制并发请求的数量
	const maxConcurrentRequests = 5
	requestCh := make(chan struct{}, maxConcurrentRequests)

	// 循环发起请求
	for i, svg := range uploadData.PageSvgs {
		go func(i int, svg string) {
			defer wg.Done() // 每个 goroutine 结束时调用 Done()

			// 获取一个并发请求的许可
			requestCh <- struct{}{}

			// 在请求结束后释放许可
			defer func() {
				<-requestCh
			}()

			png, err := _svgToPng(svg, svg2pngUrl)
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

func reviewgo(newDocument *models.Document, uploadData *UploadData, docPath string, pages []struct {
	Id string `json:"id"`
}, medias *[]Media) {
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		return
	}

	// svg2png
	pngs := svg2png(uploadData, services.GetConfig().Svg2Png.Url)

	_storage := services.GetStorageClient()

	locked := make([]models.DocumentLock, 0)
	if uploadData.DocumentText != "" {
		reviewResponse, err := reviewClient.ReviewText(uploadData.DocumentText)
		if err != nil || reviewResponse.Status != safereview.ReviewTextResultPass {
			var lockedWords string
			if wordsBytes, err := json.Marshal(reviewResponse.Words); err == nil {
				lockedWords = string(wordsBytes)
			}
			locked = append(locked, models.DocumentLock{
				DocumentId:   newDocument.Id,
				LockedReason: reviewResponse.Reason,
				LockedWords:  lockedWords,
				LockedType:   models.LockedTypeText,
			})
		}
	}
	// review pages
	if uploadData.PageSvgs != nil && len(uploadData.PageSvgs) > 0 {
		for i, page := range pages {
			png := (*pngs)[i]
			if len(png) == 0 {
				continue
			}
			path := docPath + "/page_image/" + str.IntToString(int64(i)) + ".png"
			if _, err := _storage.Bucket.PutObjectByte(path, png, ""); err != nil {
				log.Println("图片上传错误", err)
			}
			base64Str := base64.StdEncoding.EncodeToString(png)
			reviewResponse, err := (reviewClient).ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("图片审核失败", err)
				continue
			} else if reviewResponse.Status != safereview.ReviewImageResultPass {
				locked = append(locked, models.DocumentLock{
					DocumentId:   newDocument.Id,
					LockedReason: reviewResponse.Reason,
					LockedType:   models.LockedTypePage,
					LockedTarget: page.Id,
				})
			}
		}
	}

	// medias
	if medias != nil && len(*medias) > 0 {
		for _, mediaInfo := range *medias {
			base64Str := base64.StdEncoding.EncodeToString(*mediaInfo.Content)
			if len(*mediaInfo.Content) == 0 || len(base64Str) == 0 {
				continue
			}
			reviewResponse, err := (reviewClient).ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("图片审核失败", err)
				continue
			} else if reviewResponse.Status != safereview.ReviewImageResultPass {
				locked = append(locked, models.DocumentLock{
					DocumentId:   newDocument.Id,
					LockedReason: reviewResponse.Reason,
					LockedType:   models.LockedTypeMedia,
					LockedTarget: mediaInfo.Name,
				})
			}
		}
	}
	log.Println("------------------------locked----------------------------", locked)
	documentService := services.NewDocumentService()
	err := documentService.AddLockedArr(locked)
	if err != nil {
		log.Println(err)
	}
	err = documentService.DeleteAllLockedExcept(newDocument.Id, locked)
	if err != nil {
		log.Println(err)
	}
}

func review(newDocument *models.Document, uploadData *UploadData, docPath string, pages []struct {
	Id string `json:"id"`
}, medias *[]Media) {
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		return
	}
	// 没有内容
	if (uploadData.PageSvgs == nil || len(uploadData.PageSvgs) == 0) && uploadData.DocumentText == "" && (medias == nil || len(*medias) == 0) {
		return
	}
	go reviewgo(newDocument, uploadData, docPath, pages, medias)
}

// ReReviewDocument 重新审核文档接口
func ReReviewDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	documentId := c.Query("doc_id")
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}

	documentService := services.NewDocumentService()
	var document models.Document
	if err := documentService.GetById(documentId, &document); err != nil {
		response.BadRequest(c, "文档不存在")
		return
	}

	// 权限检查：只有文档所有者或有管理员权限的用户才能重新审核
	var hasPermission bool
	if document.UserId == userId {
		hasPermission = true
	} else if document.ProjectId != "" {
		// 检查项目权限，需要管理员以上权限
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		if err == nil && projectPermType != nil && (*projectPermType) >= models.ProjectPermTypeAdmin {
			hasPermission = true
		}
	}

	if !hasPermission {
		response.Forbidden(c, "无权限执行重新审核")
		return
	}

	// 获取文档内容进行审核
	err = reReviewDocumentContent(&document)
	if err != nil {
		log.Println("重新审核失败:", err)
		response.ServerError(c, "重新审核失败")
		return
	}

	response.Success(c, "重新审核已完成")
}

// reReviewDocumentContent 重新审核文档内容的内部函数
func reReviewDocumentContent(document *models.Document) error {
	// 获取配置
	config := services.GetConfig()
	reviewApiUrl := strings.Replace(config.VersionServer.Url, "generate", "review", 1)

	// 获取文档基本信息
	documentInfo, err := GetDocumentBasicInfoById(document.Id)
	if err != nil {
		return fmt.Errorf("获取文档信息失败: %w", err)
	}

	// 构建请求到版本服务器
	reqBody := map[string]interface{}{
		"documentInfo": documentInfo,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	// 发送请求到版本服务器
	resp, err := http.Post(reviewApiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("调用版本服务器失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("版本服务器返回错误: %d, %s", resp.StatusCode, string(body))
	}

	// 定义审核响应结构体
	type ReviewResp struct {
		DocumentData ExFromJson `json:"documentData"`
		DocumentText string     `json:"documentText"`
		MediasSize   uint64     `json:"mediasSize"`
		PageSvgs     []string   `json:"pageSvgs"`
	}

	// 解析文档数据
	documentData := ReviewResp{}
	err = json.Unmarshal(body, &documentData)
	if err != nil {
		return fmt.Errorf("解析文档数据失败: %w", err)
	}

	log.Printf("获取到文档数据 - DocumentText长度: %d, MediaNames数量: %d, MediasSize: %d",
		len(documentData.DocumentText), len(documentData.DocumentData.MediaNames), documentData.MediasSize)

	// 构造 UploadData 结构体
	uploadData := &UploadData{
		DocumentMeta: Data(documentData.DocumentData.DocumentMeta),
		DocumentText: documentData.DocumentText,
		MediasSize:   documentData.MediasSize,
		PageSvgs:     documentData.PageSvgs,
	}

	// 获取页面信息
	var pages []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(documentData.DocumentData.Pages, &pages); err != nil {
		log.Printf("解析页面信息失败: %v", err)
	}

	// 获取媒体文件
	var medias []Media
	_storage := services.GetStorageClient()
	docPath := document.Path

	for _, mediaName := range documentData.DocumentData.MediaNames {
		mediaPath := docPath + "/medias/" + mediaName
		mediaData, err := _storage.Bucket.GetObject(mediaPath)
		if err != nil {
			log.Printf("获取媒体文件 %s 失败: %v", mediaName, err)
			continue
		}
		medias = append(medias, Media{
			Name:    mediaName,
			Content: &mediaData,
		})
	}

	// 调用审核函数
	reviewgo(document, uploadData, docPath, pages, &medias)

	// 审核文档内容完成后，再审核评论
	reviewComment(document)

	// 审核缩略图
	reviewThumbnail(document)

	return nil
}

func reviewComment(document *models.Document) {
	// 获取审核客户端
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		log.Println("审核客户端不可用")
		return
	}

	// 获取评论服务
	commentSrv := services.GetUserCommentService()
	if commentSrv == nil {
		log.Println("评论服务不可用")
		return
	}

	// 获取文档所有评论
	comments, err := commentSrv.GetUserComment(document.Id)
	if err != nil {
		log.Printf("获取文档 %s 的评论失败: %v", document.Id, err)
		return
	}

	if len(comments) == 0 {
		log.Printf("文档 %s 没有评论需要审核", document.Id)
		return
	}

	// 文档服务用于管理锁定记录
	documentService := services.NewDocumentService()

	// 先获取当前的锁定记录，过滤出非评论类型的记录
	currentLocked, err := documentService.GetLocked(document.Id)
	if err != nil {
		log.Printf("获取文档 %s 当前锁定记录失败: %v", document.Id, err)
	}

	// 保留非评论类型的锁定记录
	nonCommentLocked := make([]models.DocumentLock, 0)
	for _, locked := range currentLocked {
		if locked.LockedType != models.LockedTypeComment {
			nonCommentLocked = append(nonCommentLocked, locked)
		}
	}

	// 用于收集需要锁定的评论
	lockedComments := make([]models.DocumentLock, 0)

	// 逐个审核评论内容
	for _, comment := range comments {
		if comment.Content == "" {
			continue // 跳过空内容的评论
		}

		// 审核评论文本内容
		reviewResponse, err := reviewClient.ReviewText(comment.Content)
		if err != nil {
			log.Printf("审核评论 %s 失败: %v", comment.CommentId, err)
			continue
		}

		// 如果审核不通过，记录锁定信息
		if reviewResponse.Status != safereview.ReviewTextResultPass {
			var lockedWords string
			if wordsBytes, err := json.Marshal(reviewResponse.Words); err == nil {
				lockedWords = string(wordsBytes)
			}

			lockedComments = append(lockedComments, models.DocumentLock{
				DocumentId:   document.Id,
				LockedType:   models.LockedTypeComment,
				LockedReason: reviewResponse.Reason,
				LockedWords:  lockedWords,
				LockedTarget: comment.CommentId,
			})

			log.Printf("评论 %s 审核不通过: %s", comment.CommentId, reviewResponse.Reason)
		}
	}

	// 删除旧的评论锁定记录，保留非评论类型的锁定记录
	err = documentService.DeleteAllLockedExcept(document.Id, nonCommentLocked)
	if err != nil {
		log.Printf("删除文档 %s 旧评论锁定记录失败: %v", document.Id, err)
	}

	// 添加新的评论锁定记录
	if len(lockedComments) > 0 {
		err = documentService.AddLockedArr(lockedComments)
		if err != nil {
			log.Printf("添加文档 %s 评论锁定记录失败: %v", document.Id, err)
		}
	}

	if len(lockedComments) > 0 {
		log.Printf("文档 %s 共审核 %d 条评论，其中 %d 条不通过",
			document.Id, len(comments), len(lockedComments))
	} else {
		log.Printf("文档 %s 的 %d 条评论均通过审核", document.Id, len(comments))
	}
}

// reviewThumbnail 审核文档缩略图
func reviewThumbnail(document *models.Document) {
	// 获取审核客户端
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		log.Println("审核客户端不可用")
		return
	}

	// 获取存储客户端
	_storage := services.GetStorageClient()
	thumbnailDir := document.Path + "/thumbnail/"

	// 列出缩略图目录下的所有文件
	objects := _storage.Bucket.ListObjects(thumbnailDir)
	var thumbnailFile string
	var thumbnailData []byte

	// 查找第一个有效的缩略图文件
	for object := range objects {
		if object.Err != nil {
			log.Printf("列出缩略图文件失败: %v", object.Err)
			continue
		}
		// 获取缩略图数据
		data, err := _storage.Bucket.GetObject(object.Key)
		if err != nil {
			log.Printf("获取缩略图 %s 失败: %v", object.Key, err)
			continue
		}
		log.Println("thumbnailData", object)
		if len(data) > 0 {
			thumbnailFile = object.Key
			thumbnailData = data
			log.Printf("找到缩略图文件: %s", thumbnailFile)
			break
		}
	}

	if thumbnailFile == "" || len(thumbnailData) == 0 {
		log.Printf("文档 %s 没有缩略图需要审核", document.Id)
		return
	}

	// 文档服务用于管理锁定记录
	documentService := services.NewDocumentService()

	// 先获取当前的锁定记录，过滤出非缩略图类型的记录
	currentLocked, err := documentService.GetLocked(document.Id)
	if err != nil {
		log.Printf("获取文档 %s 当前锁定记录失败: %v", document.Id, err)
	}

	// 保留非缩略图类型的锁定记录
	nonThumbnailLocked := make([]models.DocumentLock, 0)
	for _, locked := range currentLocked {
		if !(locked.LockedType == models.LockedTypeMedia &&
			(locked.LockedTarget == "thumbnail" || strings.Contains(locked.LockedTarget, "thumbnail"))) {
			nonThumbnailLocked = append(nonThumbnailLocked, locked)
		}
	}

	// 审核缩略图
	base64Str := base64.StdEncoding.EncodeToString(thumbnailData)
	reviewResponse, err := reviewClient.ReviewPictureFromBase64(base64Str)
	if err != nil {
		log.Printf("审核缩略图 %s 失败: %v", thumbnailFile, err)
		return
	}

	// 删除旧的缩略图锁定记录，保留非缩略图类型的锁定记录
	err = documentService.DeleteAllLockedExcept(document.Id, nonThumbnailLocked)
	if err != nil {
		log.Printf("删除文档 %s 旧缩略图锁定记录失败: %v", document.Id, err)
	}

	// 如果审核不通过，记录锁定信息
	if reviewResponse.Status != safereview.ReviewImageResultPass {
		lockedThumbnail := models.DocumentLock{
			DocumentId:   document.Id,
			LockedType:   models.LockedTypeMedia,
			LockedReason: reviewResponse.Reason,
			LockedTarget: "thumbnail",
		}

		err = documentService.AddLocked(&lockedThumbnail)
		if err != nil {
			log.Printf("添加文档 %s 缩略图锁定记录失败: %v", document.Id, err)
		} else {
			log.Printf("缩略图 审核不通过: %s", reviewResponse.Reason)
		}
	} else {
		log.Printf("文档 %s 的缩略图审核通过", document.Id)
	}
}
