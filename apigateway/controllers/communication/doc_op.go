package communication

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/apigateway/common/docop"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
)

func OpenDocOpTunnel(clientWs *websocket.Ws, clientCmdData CmdData, serverCmd ServerCmd, data Data) *Tunnel {
	clientCmdDataData, ok := clientCmdData["data"].(map[string]any)
	documentIdStr, ok1 := clientCmdDataData["document_id"].(string)
	userId, ok2 := data["userId"].(int64)
	if !ok || !ok1 || documentIdStr == "" || !ok2 || userId <= 0 {
		serverCmd.Message = "通道建立失败，参数错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，参数错误", ok, ok1, ok2, documentIdStr, userId)
		return nil
	}
	versionId, _ := clientCmdDataData["version_id"].(string)
	previousCmdId, _ := clientCmdDataData["previous_cmd_id"].(string)

	// 获取文档信息
	documentId := str.DefaultToInt(documentIdStr, 0)
	if documentId <= 0 {
		serverCmd.Message = "通道建立失败，documentId错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，documentId错误", documentId)
		return nil
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(documentId, &document) != nil {
		serverCmd.Message = "通道建立失败，文档不存在"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，文档不存在", documentId)
		return nil
	}
	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
		serverCmd.Message = "通道建立失败"
		if err != nil {
			serverCmd.Message += "，无权限"
		}
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，权限校验错误", err, permType)
		return nil
	}
	// 验证文档版本信息
	if versionId != "" {
		var documentVersion models.DocumentVersion
		if err := documentService.DocumentVersionService.Get(&documentVersion, "document_id = ? and version_id = ?", documentId, versionId); err != nil {
			serverCmd.Message = "通道建立失败，文档版本错误"
			_ = clientWs.WriteJSON(&serverCmd)
			log.Println("document ws建立失败，文档版本不存在", documentId, versionId)
			return nil
		}
	}
	if !document.LockedAt.IsZero() && document.UserId != userId {
		serverCmd.Message = "通道建立失败，审核不通过"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，审核不通过", documentId)
		return nil
	}

	docopUrl := docop.GetDocumentUrlRetry(documentIdStr)
	if docopUrl == "" {
		serverCmd.Message = "通道建立失败"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，文档服务错误", documentId)
		return nil
	}

	serverWs, err := websocket.NewClient(docopUrl, nil)
	if err != nil {
		serverCmd.Message = "通道建立失败"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败", err)
		return nil
	}
	if err := serverWs.WriteJSON(Data{
		"documentId":    documentIdStr,
		"userId":        str.IntToString(userId),
		"versionId":     versionId,
		"previousCmdId": previousCmdId,
	}); err != nil {
		serverCmd.Message = "通道建立失败"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败（鉴权）", err)
		return nil
	}

	tunnelId := uuid.New().String()
	tunnel := &Tunnel{
		Id:     tunnelId,
		Server: serverWs,
		Client: clientWs,
	}
	// 转发客户端数据到服务端
	tunnel.ReceiveFromClient = func(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd) error {
		if permType >= models.PermTypeEditable {
			err := tunnel.Server.WriteMessage(websocket.MessageType(tunnelDataType), data)
			if err != nil {
				log.Println("数据发送失败", err)
				return errors.New("数据发送失败")
			}
		} else {
			return errors.New("无权限")
		}
		return nil
	}
	// 转发服务端数据到客户端
	go tunnel.DefaultServerToClient()

	return tunnel
}
