package communication

import (
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/services"
)

func GetUserDocumentInfo(userId int64, documentId int64, permType models.PermType) (*services.DocumentInfoQueryRes, string) {

	// permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	// if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
	// 	permType = models.PermTypeNone
	// }
	result := services.NewDocumentService().GetDocumentInfoByDocumentAndUserId(documentId, userId, permType)
	if result == nil {
		return nil, "文档不存在"
	} else if !result.Document.LockedAt.IsZero() && result.Document.UserId != userId {
		return nil, "审核不通过"
	} else {
		return result, ""
	}
}
