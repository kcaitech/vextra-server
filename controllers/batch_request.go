package controllers

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/response"
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
	Reqid int64            `json:"reqid"`
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

type BatchResponseItem struct {
	Reqid int64       `json:"reqid"`
	Error string      `json:"error,omitempty"`
	Sha1  string      `json:"sha1,omitempty"`
	Data  interface{} `json:"data"`
}

func batch_request(c *gin.Context, router *gin.Engine) {
	var batchRequests []BatchRequestItem
	if err := c.ShouldBindJSON(&batchRequests); err != nil {
		log.Println("args", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 复制原始请求的头部信息
	originalHeaders := c.Request.Header
	results := make([]BatchResponseItem, len(batchRequests))
	for i, req := range batchRequests {
		// 创建一个新的 Gin 上下文
		respWriter := ResponseWriter{Body: &bytes.Buffer{}, StatusCode: http.StatusOK, header: http.Header{}}
		newCtx, _ := gin.CreateTestContext(&respWriter)

		joinUrl := path.Join("/api", req.Data.Url)
		if strings.HasSuffix(req.Data.Url, "/") && !strings.HasSuffix(joinUrl, "/") { // 需要保留最后的"/"
			joinUrl += "/"
		}
		// 设置请求方法和路径
		newCtx.Request = &http.Request{
			Method: strings.ToUpper(req.Data.Method),
			URL:    &url.URL{Path: joinUrl},
		}

		newCtx.Request.Header = make(http.Header)
		// 复制原始请求的头部信息到子请求
		for k, v := range originalHeaders {
			newCtx.Request.Header[k] = v
		}

		if req.Data.Params != nil {
			// for key, value := range req.Data.Params {
			// 	newCtx.Params = append(newCtx.Params, gin.Param{Key: key, Value: fmt.Sprintf("%v", value)})
			// }
			// 设置查询参数
			query := url.Values{}
			for key, value := range req.Data.Params {
				var str string
				if _str, ok := value.(string); ok {
					str = _str
				} else {
					v, _ := json.Marshal(value)
					str = string(v)
				}
				query.Add(key, str)
			}
			newCtx.Request.URL.RawQuery = query.Encode()
		}

		if req.Data.Data != nil {
			var data, _ = json.Marshal(req.Data.Data)
			if data != nil {
				body := bytes.NewReader(data)
				newCtx.Request.Body = io.NopCloser(body)
				newCtx.Request.Header["Content-Type"] = []string{"application/json"}
			}
		}

		newCtx.Request.Header["Accept-Encoding"] = []string{""} // 不要gzip压缩
		// 执行路由处理函数
		router.HandleContext(newCtx)

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
			results[i] = BatchResponseItem{Reqid: req.Reqid, Sha1: sha1}
			continue
		}

		if !isOk {
			log.Println("not ok, data:", string(data), ", status: ", respWriter.StatusCode)
			results[i] = BatchResponseItem{Reqid: req.Reqid, Error: string(data)}
		} else if err := json.Unmarshal(data, &result); err != nil {
			log.Println("unmarshal", err, ", data:", string(data))
			results[i] = BatchResponseItem{Reqid: req.Reqid, Error: err.Error()}
		} else if sha1 != "" {
			results[i] = BatchResponseItem{Reqid: req.Reqid, Data: result, Sha1: sha1}
		} else {
			results[i] = BatchResponseItem{Reqid: req.Reqid, Data: result}
		}
	}

	response.Success(c, &results)
}

func BatchRequestHandler(router *gin.Engine) func(c *gin.Context) {
	return func(c *gin.Context) {
		batch_request(c, router)
	}
}
