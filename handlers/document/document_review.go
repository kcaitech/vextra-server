package document

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"os"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
)

func reviewgo(newDocument models.Document, uploadData *VersionResp, docPath string, pages []struct {
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
	tmp_dir := services.GetConfig().SafeReview.TmpPngDir + "/" + newDocument.Id
	// review pages
	if len(uploadData.PagePngs) > 0 {
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
				log.Println("读取png文件失败", err, tmp_dir+"/"+pagePng)
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
	if (len(uploadData.PagePngs) == 0) && uploadData.DocumentText == "" && (medias == nil || len(*medias) == 0) {
		return
	}
	go reviewgo(*newDocument, uploadData, docPath, pages, medias)
}
