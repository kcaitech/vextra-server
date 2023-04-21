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
)

type uploadData struct {
	Code string `json:"code"`
	Data struct {
		DocumentMeta     map[string]interface{}   `json:"document_meta"`
		Pages            json.RawMessage          `json:"pages"`
		PageRefartboards [][]string               `json:"page_refartboards"`
		PageRefsyms      [][]string               `json:"page_refsyms"`
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

	docId := uuid.New().String()

	for {
		messageType, reader, err := conn.NextReader()
		if err != nil {
			connError()
			break
		}

		if messageType == websocket.TextMessage {
			_data, err := io.ReadAll(reader)
			if err != nil {
				connError()
				break
			}

			var data uploadData
			if err := json.Unmarshal(_data, &data); err != nil {
				dataFormatError()
				break
			}
			var pages []struct {
				Id string `json:"id"`
			}
			if err := json.Unmarshal(data.Data.Pages, &pages); err != nil {
				dataFormatError()
				break
			}
			var pagesRaw []json.RawMessage
			if err := json.Unmarshal(data.Data.Pages, &pagesRaw); err != nil {
				dataFormatError()
				break
			}

			if len(pages) != len(data.Data.PageRefartboards) || len(pages) != len(data.Data.PageRefsyms) {
				dataFormatError()
				break
			}

			for i := 0; i < len(pages); i++ {
				pageId := pages[i].Id
				path := docId + "/pages/" + pageId + ".json"
				path = docId + "/page-symrefs/" + pageId + ".json"
				path = docId + "/page-artboardrefs/" + pageId + ".json"
				log.Println(path)
				// 写入对象存储
			}

			path := docId + "/document-meta.json"
			log.Println(path)
			// 写入对象存储
		}

		if messageType == websocket.CloseMessage {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			break
		}
	}
}
