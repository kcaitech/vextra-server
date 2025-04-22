package ws

import (
	"encoding/json"
	"log"

	"kcaitech.com/kcserver/utils/websocket"

	document "kcaitech.com/kcserver/handlers/document"
)

type Export struct {
	DocumentMeta Data            `json:"document_meta"`
	Pages        json.RawMessage `json:"pages"`
	MediaNames   []string        `json:"media_names"`
}

type DocData struct {
	Id         string
	ProjectId  string
	Export     *Export
	Medias     []document.Media
	MediasSize uint64
}

type docUploadServe struct {
	ws     *websocket.Ws
	userId string

	// cache // todo cache in redis
	data *DocData
}

func NewDocUploadServe(ws *websocket.Ws, userId string) *docUploadServe {

	// 权限校验
	// var permType models.PermType
	// if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
	// 	log.Println("NO comment perm", err, permType)
	// 	return nil
	// }

	serv := docUploadServe{
		ws:     ws,
		userId: userId,
	}
	serv.start()
	return &serv
}

func (serv *docUploadServe) start() {

}

func (serv *docUploadServe) close() {

}

func (serv *docUploadServe) handle(data *TransData, binaryData *([]byte)) {

	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId
	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Msg = msg
		if err != nil {
			log.Println(msg, *err)
		} else {
			log.Println(msg)
		}
		_ = serv.ws.WriteJSON(serverData)
	}

	type UploadHeader struct {
		DocumentId string  `json:"document_id"`
		ProjectId  string  `json:"project_id,omitempty"`
		Export     *Export `json:"export,omitempty"`
		Commit     bool    `json:"commit,omitempty"`
		Media      string  `json:"media,omitempty"`
	}

	uploadHeader := &UploadHeader{}
	err := json.Unmarshal([]byte(data.Data), uploadHeader)
	if err != nil {
		msgErr("数据错误", &serverData, &err)
		return
	}

	if serv.data == nil || serv.data.Id != uploadHeader.DocumentId {
		log.Println("uploading", uploadHeader.DocumentId)
		serv.data = &DocData{
			Id:        uploadHeader.DocumentId,
			ProjectId: uploadHeader.ProjectId,
			Medias:    []document.Media{},
		}
	}

	if uploadHeader.Export != nil {
		serv.data.Export = uploadHeader.Export
		_ = serv.ws.WriteJSON(serverData)
		return
	}
	if uploadHeader.Media != "" && binaryData != nil {
		media := document.Media{
			Name:    uploadHeader.Media,
			Content: binaryData,
		}
		serv.data.Medias = append(serv.data.Medias, media)
		serv.data.MediasSize += uint64(len(*binaryData))
		_ = serv.ws.WriteJSON(serverData)
		return
	}
	if uploadHeader.Commit && serv.data != nil && serv.data.Export != nil {
		log.Println("uploading commit", uploadHeader.DocumentId)
		header := document.Header{
			UserId:    (serv.userId),
			ProjectId: serv.data.ProjectId,
		}

		uploadData := document.UploadData{
			DocumentMeta: document.Data(serv.data.Export.DocumentMeta),
			Pages:        serv.data.Export.Pages,
			MediaNames:   serv.data.Export.MediaNames,
			MediasSize:   serv.data.MediasSize,
		}

		resp := document.Response{}
		document.UploadDocumentData(&header, &uploadData, &serv.data.Medias, &resp)

		if resp.Status == document.ResponseStatusSuccess {
			retData, err := json.Marshal(resp.Data)
			if err != nil {
				log.Println("resp.Data错误??", err)
			}
			serverData.Data = string(retData)
			_ = serv.ws.WriteJSON(serverData)
			serv.data = nil // 已上传成功
		} else {
			msgErr(resp.Message, &serverData, &err)
		}
		return
	}
	msgErr("数据错误", &serverData, nil)
}
