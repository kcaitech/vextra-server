package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/storage"
)

type uploadData struct {
	Code string `json:"code"`
	Data struct {
		DocumentMeta     json.RawMessage          `json:"document_meta"`
		Pages            json.RawMessage          `json:"pages"`
		PageRefartboards []json.RawMessage        `json:"page_refartboards"`
		PageRefsyms      []json.RawMessage        `json:"page_refsyms"`
		Artboards        []map[string]interface{} `json:"artboards"`
		ArtboardsRefsyms [][]string               `json:"artboard_refsyms"`
		Symbols          []map[string]interface{} `json:"symbols"`
		Medias           []string                 `json:"medias"`
	} `json:"data"`
}

// UploadHandler 上传文档
func UploadHandler(c *gin.Context) {
	//userId, err := auth.GetUserId(c)
	//if err != nil {
	//	response.Unauthorized(c)
	//	return
	//}
	//
	//userService := services.NewUserService()
	//_, err = userService.GetUser(userId)
	//if err != nil {
	//	response.BadRequest(c, err.Error())
	//	return
	//}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer conn.Close()

	connError := func() {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			"ws连接异常",
		))
	}

	dataFormatError := func() {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			"数据格式错误",
		))
	}

	uploadError := func(err error) {
		log.Println("对象上传错误：" + err.Error())
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			"对象上传错误",
		))
	}

	docId := uuid.New().String()

	for {
		messageType, reader, err := conn.NextReader()
		if err != nil {
			connError()
			return
		}

		if messageType == websocket.TextMessage {
			_data, err := io.ReadAll(reader)
			if err != nil {
				connError()
				return
			}

			var data uploadData
			if err := json.Unmarshal(_data, &data); err != nil {
				dataFormatError()
				return
			}
			var pages []struct {
				Id string `json:"id"`
			}
			if err := json.Unmarshal(data.Data.Pages, &pages); err != nil {
				dataFormatError()
				return
			}
			var pagesRaw []json.RawMessage
			if err := json.Unmarshal(data.Data.Pages, &pagesRaw); err != nil {
				dataFormatError()
				return
			}

			if len(pages) != len(data.Data.PageRefartboards) || len(pages) != len(data.Data.PageRefsyms) {
				dataFormatError()
				return
			}

			for i := 0; i < len(pages); i++ {
				pageId := pages[i].Id
				path := docId + "/pages/" + pageId + ".json"
				if _, err = storage.Bucket.PubObjectByte(path, pagesRaw[i]); err != nil {
					uploadError(err)
					return
				}
				path = docId + "/page-symrefs/" + pageId + ".json"
				if _, err = storage.Bucket.PubObjectByte(path, data.Data.PageRefsyms[i]); err != nil {
					uploadError(err)
					return
				}
				path = docId + "/page-artboardrefs/" + pageId + ".json"
				if _, err = storage.Bucket.PubObjectByte(path, data.Data.PageRefartboards[i]); err != nil {
					uploadError(err)
					return
				}
			}

			path := docId + "/document-meta.json"
			if _, err = storage.Bucket.PubObjectByte(path, data.Data.DocumentMeta); err != nil {
				uploadError(err)
				return
			}
		}

		if messageType == websocket.CloseMessage {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}
