/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"kcaitech.com/kcserver/utils/websocket"

	"kcaitech.com/kcserver/handlers/common"
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
	Medias     []common.Media
	MediasSize uint64
}

type docUploadServe struct {
	ws     *websocket.Ws
	userId string
	data   *DocData
}

func NewDocUploadServe(ws *websocket.Ws, userId string) *docUploadServe {
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
			Medias:    []common.Media{},
		}
	}

	if uploadHeader.Export != nil {
		serv.data.Export = uploadHeader.Export
		_ = serv.ws.WriteJSON(serverData)
		return
	}
	if uploadHeader.Media != "" && binaryData != nil {
		// 去除失败重传的图片
		for _, media := range serv.data.Medias {
			if media.Name == uploadHeader.Media {
				log.Println("uploading media already exists", uploadHeader.Media, " size:", uint64(len(*binaryData)))
				_ = serv.ws.WriteJSON(serverData)
				return
			}
		}
		media := common.Media{
			Name:    uploadHeader.Media,
			Content: binaryData,
		}
		serv.data.Medias = append(serv.data.Medias, media)
		serv.data.MediasSize += uint64(len(*binaryData))
		log.Println("uploading media", uploadHeader.Media, " size:", uint64(len(*binaryData)),
			"already uploaded media count:", len(serv.data.Medias), " total size:", serv.data.MediasSize)
		_ = serv.ws.WriteJSON(serverData)
		return
	}
	if uploadHeader.Commit && serv.data != nil && serv.data.Export != nil {
		log.Println("uploading commit", uploadHeader.DocumentId)
		// header := document.Header{
		// 	UserId:    (serv.userId),
		// 	ProjectId: serv.data.ProjectId,
		// }

		uploadData := common.VersionResp{
			DocumentData: common.ExFromJson{
				DocumentMeta: common.Data(serv.data.Export.DocumentMeta),
				Pages:        serv.data.Export.Pages,
				MediaNames:   serv.data.Export.MediaNames,
			},
			MediasSize: serv.data.MediasSize,
		}

		resp := common.Response{}
		common.UploadNewDocumentData(serv.userId, serv.data.ProjectId, &uploadData, &serv.data.Medias, &resp)

		if resp.Code == http.StatusOK {
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
