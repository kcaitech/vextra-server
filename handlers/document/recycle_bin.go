package document

import (
	"errors"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

// GetUserRecycleBinDocumentList 获取用户回收站文档列表
func GetUserRecycleBinDocumentList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := c.Query("project_id")
	recycleBinList := services.NewDocumentService().FindRecycleBinByUserId(userId, projectId)
	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range *recycleBinList {
		userIds = append(userIds, item.Document.UserId)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	result := make([]services.RecycleBinQueryResItem, 0)
	for _, item := range *recycleBinList {
		userId := item.Document.UserId
		userInfo, exists := userMap[userId]
		deleteById := item.Document.DeleteBy
		if deleteById != "" {
			deleteUserInfo, deleteExists := userMap[deleteById]
			if deleteExists {
				item.DeleteUser = &models.UserProfile{
					Id:       deleteUserInfo.UserID,
					Nickname: deleteUserInfo.Nickname,
					Avatar:   deleteUserInfo.Avatar,
				}
			}
		}
		if exists {
			item.User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.Avatar,
			}
			result = append(result, item)
		}
	}
	response.Success(c, result)
}

type RestoreUserRecycleBinDocumentReq struct {
	DocId string `json:"doc_id" binding:"required"`
}

// RestoreUserRecycleBinDocument 恢复用户回收站内的某份文档
func RestoreUserRecycleBinDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req RestoreUserRecycleBinDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := (req.DocId)
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	document := models.Document{}
	if documentService.Get(&document, "id = ? and deleted_at is not null", documentId, &services.Unscoped{}) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == "" {
		if document.UserId != userId {
			response.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		// 管理员以上权限或文档创建者且可编辑权限
		if !(err == nil && projectPermType != nil && ((*projectPermType) >= models.ProjectPermTypeAdmin || ((*projectPermType) == models.ProjectPermTypeEditable && document.UserId == userId))) {
			response.Forbidden(c, "")
			return
		}
	}
	if _, err := documentService.UpdateColumns(
		map[string]any{"deleted_at": nil},
		"id = ? and deleted_at is not null", documentId,
		&services.Unscoped{},
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// DeleteUserRecycleBinDocument 彻底删除用户回收站内的某份文档
func DeleteUserRecycleBinDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := c.Query("doc_id")
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	document := models.Document{}
	if documentService.Get(&document, "id = ? and deleted_at is not null", documentId, &services.Unscoped{}) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == "" {
		if document.UserId != userId {
			response.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		// 管理员以上权限或文档创建者且可编辑权限
		if !(err == nil && projectPermType != nil && ((*projectPermType) >= models.ProjectPermTypeAdmin || ((*projectPermType) == models.ProjectPermTypeEditable && document.UserId == userId))) {
			response.Forbidden(c, "")
			return
		}
	}
	_, err = documentService.HardDelete("id = ? and deleted_at is not null", documentId)
	if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "更新错误")
		return
	}
	// todo 删除oss文件
	response.Success(c, "")
}
