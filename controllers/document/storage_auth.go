package document

import (
	"github.com/gin-gonic/gin"
	"log"
	"kcaitech.com/kcserver/common/gin/response"
)

// CheckStorageAuth storage授权
func CheckStorageAuth(c *gin.Context) {
	//userId, err := auth.GetUserId(c)
	//if err != nil {
	//	response.Unauthorized(c)
	//	return
	//}

	log.Println("CheckStorageAuth", c.Query("t"))

	response.Success(c, nil)
}
