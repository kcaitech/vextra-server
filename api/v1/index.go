package v1

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
)

func LoadRoutes(apiGroup *gin.RouterGroup) {
	loadWsRoutes(apiGroup)    // 单独鉴权
	loadLoginRoutes(apiGroup) // 从refreshToken获取信息
	loadAccessRoutes(apiGroup)
	apiGroup.Use(services.GetKCAuthClient().AuthRequired())
	apiGroup.Use(common.Sha1SaveData)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadShareRoutes(apiGroup)
	loadTeamRoutes(apiGroup)
	loadFeedbackRoutes(apiGroup)
}
