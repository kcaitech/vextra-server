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
	"os"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
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

func reviewgo(newDocument *models.Document, uploadData *VersionResp, docPath string, pages []struct {
	Id string `json:"id"`
}, medias *[]Media) {
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		return
	}

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
	tmp_dir := services.GetConfig().TmpDir + "/" + newDocument.Id
	// review pages
	if uploadData.PagePngs != nil && len(uploadData.PagePngs) > 0 {
		for _, page := range pages {
			pagePng := page.Id + ".png"

			png := ""
			for _, pagePngName := range uploadData.PagePngs {
				if pagePngName == pagePng {
					png = pagePngName
					break
				}
			}
			if len(png) == 0 {
				continue
			}

			path := docPath + "/page_image/" + page.Id + ".png"

			// 读取png文件
			pngBytes, err := os.ReadFile(tmp_dir + "/" + pagePng)
			if err != nil {
				log.Println("读取png文件失败", err)
				continue
			}

			if _, err := _storage.Bucket.PutObjectByte(path, pngBytes, ""); err != nil {
				log.Println("图片上传错误", err)
			}
			base64Str := base64.StdEncoding.EncodeToString(pngBytes)
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
	// 清空临时目录
	os.RemoveAll(tmp_dir)

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

func review(newDocument *models.Document, uploadData *VersionResp, docPath string, pages []struct {
	Id string `json:"id"`
}, medias *[]Media) {
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		return
	}
	// 没有内容
	if (uploadData.PagePngs == nil || len(uploadData.PagePngs) == 0) && uploadData.DocumentText == "" && (medias == nil || len(*medias) == 0) {
		return
	}
	go reviewgo(newDocument, uploadData, docPath, pages, medias)
}
