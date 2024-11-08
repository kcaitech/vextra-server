package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type ops = map[string]interface{}

type BatchRequestItem struct {
	ops
	Method string `json:"method"`
	Url    string `json:"url"`
	Params string `json:"params"`
}

// 自定义 UnmarshalJSON 方法
func (bri *BatchRequestItem) UnmarshalJSON(data []byte) error {
	type Alias BatchRequestItem
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(bri),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	return nil
}

func batch_request(c *gin.Context, router *gin.Engine) {
	var batchRequests []BatchRequestItem
	if err := c.ShouldBindJSON(&batchRequests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	results := make([]map[string]interface{}, len(batchRequests))
	for i, req := range batchRequests {
		// 创建一个新的 Gin 上下文
		newCtx, _ := gin.CreateTestContext(nil)

		// 设置请求方法和路径
		newCtx.Request = &http.Request{
			Method: req.Method,
			URL:    &url.URL{Path: req.Url},
		}

		// 设置请求体
		// if req.ops.Body != "" {
		// 	body := strings.NewReader(req.Body)
		// 	newCtx.Request.Body = ioutil.NopCloser(body)
		// }

		// 执行路由处理函数
		router.ServeHTTP(newCtx.Writer, newCtx.Request)

		// 解析响应
		var result map[string]interface{}
		// if err := json.Unmarshal(newCtx.Writer.Body.Bytes(), &result); err != nil {
		// 	result = map[string]interface{}{"error": err.Error()}
		// }
		results[i] = result
	}

	c.JSON(http.StatusOK, results)
}

func BatchRequestHandler(router *gin.Engine) func(c *gin.Context) {
	return func(c *gin.Context) {
		batch_request(c, router)
	}
}
