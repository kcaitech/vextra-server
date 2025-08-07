/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package common

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SHA1ResponseWriter 用于拦截响应数据
type SHA1ResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *SHA1ResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *SHA1ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// sha1base64 计算数据的SHA1哈希值并转换为Base64字符串
func sha1base64(data []byte) string {
	hash := sha1.Sum(data)
	return base64.URLEncoding.EncodeToString(hash[:])
}

func Sha1SaveData(c *gin.Context) {
	clientSha1 := c.Query("sha1")

	// 创建响应拦截器
	writer := &SHA1ResponseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
	c.Writer = writer

	// 继续处理请求
	c.Next()

	// 获取响应数据
	responseData := writer.body.Bytes()

	// 如果没有响应数据或数据太小，直接返回原始响应
	if len(responseData) == 0 || len(responseData) <= 128 {
		writer.ResponseWriter.Write(responseData)
		return
	}

	// 计算响应数据的SHA1
	currentSha1 := sha1base64(responseData)

	// 如果SHA1相同，只返回包含SHA1的响应
	if clientSha1 == currentSha1 {
		// 不需要传递data，因为客户端已经有缓存的数据
		// 只需要返回SHA1确认缓存仍然有效
		writer.ResponseWriter.Header().Set("Content-Type", "application/json")
		writer.ResponseWriter.WriteHeader(writer.statusCode)

		onlysha1Response := Response{
			Code:    http.StatusOK,
			Message: "",
			Data:    nil,
			Sha1:    currentSha1,
		}

		json.NewEncoder(writer.ResponseWriter).Encode(onlysha1Response)
		return
	}

	// SHA1不同，解析原始响应并添加SHA1
	var originalResponse Response
	if err := json.Unmarshal(responseData, &originalResponse); err != nil {
		// 如果解析失败，直接返回原始响应
		writer.ResponseWriter.Write(responseData)
		return
	}

	// 添加SHA1到响应中
	originalResponse.Sha1 = currentSha1

	// 返回修改后的响应
	writer.ResponseWriter.Header().Set("Content-Type", "application/json")
	writer.ResponseWriter.WriteHeader(writer.statusCode)
	json.NewEncoder(writer.ResponseWriter).Encode(originalResponse)
}
