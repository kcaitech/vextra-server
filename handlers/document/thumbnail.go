package document

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
)

func GetDocumentThumbnailAccessKey(c *gin.Context) {
	docId := c.Query("doc_id")
	key, code, err := common.GetDocumentThumbnailAccessKey(c, docId, services.GetStorageClient())
	if err == nil {
		common.Success(c, key)
	} else if code == http.StatusUnauthorized {
		common.Unauthorized(c)
	} else {
		common.ServerError(c, err.Error())
	}
}
