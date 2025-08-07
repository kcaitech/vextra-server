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
	"context"
	"fmt"
	"log"

	"kcaitech.com/kcserver/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/redis"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/websocket"
)

type VersionServe struct {
	ws   *websocket.Ws
	quit chan struct{}
	// isready bool
	genSId func() string
	redis  *redis.RedisDB
}

func NewVersionServe(ws *websocket.Ws, userId string, documentId string, genSId func() string) *VersionServe {
	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
		log.Println("Insufficient permissions to create version service", err, permType)
		return nil
	}

	serv := VersionServe{
		ws: ws,
		// isready: false,
		genSId: genSId,
		quit:   make(chan struct{}),
		redis:  services.GetRedisDB(),
	}

	serv.start(documentId)
	// serv.isready = true
	return &serv
}

func (serv *VersionServe) start(documentId string) {

	// documentIdStr := str.IntToString(documentId)
	// 监控评论变化
	go func() {
		// defer tunnelServer.Close()
		pubsub := serv.redis.Client.Subscribe(context.Background(), fmt.Sprintf("%s%s", common.RedisKeyDocumentVersion, documentId))
		defer pubsub.Close()
		channel := pubsub.Channel()
		for {
			select {
			case v, ok := <-channel:
				if !ok {
					break
				}
				serv.send(v.Payload)
			case <-serv.quit:
				return
			}
		}
	}()
}

func (serv *VersionServe) close() {
	close(serv.quit)
}

func (serv *VersionServe) handle(data *TransData, binaryData *([]byte)) {
}

func (serv *VersionServe) send(data string) {
	sid := serv.genSId()
	serverData := TransData{
		Type:   DataTypes_GenerateVersion,
		DataId: sid,
		Data:   data,
	}
	if err := serv.ws.WriteJSONLock(true, &serverData); err != nil {
		log.Println("version, send data fail", err)
		return
	}
}
