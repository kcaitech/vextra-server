package controllers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type BatchRequestData struct {
	Method string                 `json:"method"`
	Url    string                 `json:"url"`
	Params map[string]interface{} `json:"params,omitempty"` // Get
	Data   map[string]interface{} `json:"data,omitempty"`   // Post
}

// 自定义 UnmarshalJSON 方法
// func (bri *BatchRequestData) UnmarshalJSON(data []byte) error {
// 	type Alias BatchRequestData
// 	aux := &struct {
// 		*Alias
// 	}{
// 		Alias: (*Alias)(bri),
// 	}
// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}
// 	return nil
// }

type BatchRequestItem struct {
	Reqid uint64           `json:"reqid"`
	Sha1  string           `json:"sha1,omitempty"`
	Data  BatchRequestData `json:"data"`
}

type ResponseWriter struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
	header     http.Header
}

func (w *ResponseWriter) Header() http.Header {
	return w.header
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	return w.Body.Write(b)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func sha1base64(data []byte) string {
	// 计算 SHA-1 哈希值
	hash := sha1.Sum(data)
	// 将哈希值转换为 Base64 URL 安全的字符串
	return base64.URLEncoding.EncodeToString(hash[:])
}

func batch_request(c *gin.Context, router *gin.Engine) {
	var batchRequests []BatchRequestItem
	if err := c.ShouldBindJSON(&batchRequests); err != nil {
		log.Println("args", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	results := make([]map[string]interface{}, len(batchRequests))
	for i, req := range batchRequests {
		// 创建一个新的 Gin 上下文
		respWriter := ResponseWriter{Body: &bytes.Buffer{}, StatusCode: http.StatusOK, header: http.Header{}}
		newCtx, _ := gin.CreateTestContext(&respWriter)

		// copy header
		for key, val := range c.Request.Header {
			for _, v := range val {
				newCtx.Request.Header.Set(key, v)
			}
		}

		path := req.Data.Url
		var method = strings.ToLower(req.Data.Method)
		if method == "get" && req.Data.Params != nil {
			queryParams := url.Values{}
			for key, value := range req.Data.Params {
				queryParams.Add(key, fmt.Sprintf("%v", value))
			}
			if len(queryParams) > 0 {
				path += "?" + queryParams.Encode()
			}
		} else if method == "post" && req.Data.Data != nil {
			var data, _ = json.Marshal(req.Data.Data)
			if data != nil {
				body := bytes.NewReader(data)
				newCtx.Request.Body = io.NopCloser(body)
				newCtx.Request.Header.Set("Content-Type", "application/json")
			}
		}

		// 设置请求方法和路径
		newCtx.Request = &http.Request{
			Method: req.Data.Method,
			URL:    &url.URL{Path: path},
		}

		// 执行路由处理函数
		router.ServeHTTP(newCtx.Writer, newCtx.Request)

		// 解析响应
		var result map[string]interface{}
		var isOk = respWriter.StatusCode == http.StatusOK
		data := respWriter.Body.Bytes()
		var sha1 string
		if len(data) > 128 {
			sha1 = sha1base64(data)
		}
		if req.Sha1 != "" && sha1 != "" && sha1 == req.Sha1 {
			// 没变化，不需要返回data
			results[i] = map[string]interface{}{"reqid": req.Reqid, "sha1": sha1}
			continue
		}

		if !isOk {
			if data != nil {
				log.Println("not ok, data:", data, ", status: ", respWriter.StatusCode)
			} else {
				log.Println("not ok, status: ", respWriter.StatusCode)
			}
			results[i] = map[string]interface{}{"reqid": req.Reqid, "error": data}
		} else if err := json.Unmarshal(data, &result); err != nil {
			log.Println("unmarshal", err)
			results[i] = map[string]interface{}{"reqid": req.Reqid, "error": err.Error()}
		} else if sha1 != "" {
			results[i] = map[string]interface{}{"reqid": req.Reqid, "data": result, "sha1": sha1}
		} else {
			results[i] = map[string]interface{}{"reqid": req.Reqid, "data": result}
		}
	}

	c.JSON(http.StatusOK, results)
}

func BatchRequestHandler(router *gin.Engine) func(c *gin.Context) {
	return func(c *gin.Context) {
		batch_request(c, router)
	}
}
