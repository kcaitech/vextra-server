package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

// GetUserDocumentSharesList 获取用户的共享文档列表
func GetUserDocumentSharesList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	response.Success(c, services.NewDocumentService().FindSharesByUserId(userId))
}

// DeleteUserShare 退出共享
func DeleteUserShare(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	permissionId := str.DefaultToInt(c.Query("share_id"), 0)
	if permissionId <= 0 {
		response.BadRequest(c, "参数错误：share_id")
		return
	}
	if err := services.NewDocumentService().DocumentPermissionService.Delete(
		"grantee_type = ? and grantee_id = ? and id = ?", models.GranteeTypeExternal, userId, permissionId,
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "删除错误")
		return
	}
	response.Success(c, nil)
}

type SetDocumentShareTypeReq struct {
	DocId   string         `json:"doc_id" binding:"required"`
	DocType models.DocType `json:"doc_type"`
}

// SetDocumentShareType 设置文档分享类型
func SetDocumentShareType(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req SetDocumentShareTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(req.DocId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	docType := req.DocType
	if docType > models.DocTypePublicEditable {
		response.BadRequest(c, "参数错误：doc_type")
		return
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if err := documentService.Get(&document, "id = ? and user_id = ?", documentId, userId); err != nil {
		if err == services.ErrRecordNotFound {
			response.Unauthorized(c)
		} else {
			response.Fail(c, "查询错误")
		}
		return
	}
	document.DocType = docType
	if err := documentService.UpdatesById(documentId, &document); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// GetDocumentSharesList 获取分享列表
func GetDocumentSharesList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := str.DefaultToInt(c.Query("doc_id"), 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	var count int64
	if err := documentService.Count(&count, "id = ? and user_id = ?", documentId, userId); err != nil || count <= 0 {
		response.Forbidden(c, "")
		return
	}
	response.Success(c, documentService.FindSharesByDocumentId(documentId))
}

// SetDocumentSharePermission 修改分享权限
func SetDocumentSharePermission(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	permissionId := str.DefaultToInt(c.Query("share_id"), 0)
	if permissionId <= 0 {
		response.BadRequest(c, "参数错误：share_id")
		return
	}
	permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	if err := services.NewDocumentService().DocumentPermissionService.UpdateColumns(
		map[string]any{"document_permission.perm_type": permType},
		"document_permission.resource_type = ? and document_permission.id", models.ResourceTypeDoc, permissionId,
		services.JoinArgs{Join: "inner join document on document.id = document_permission.resource_id", Args: nil},
		services.WhereArgs{Query: "document.user_id = ?", Args: []any{userId}},
	); err != nil {
		if err == services.ErrRecordNotFound {
			response.Unauthorized(c)
		} else {
			response.Fail(c, "更新错误")
		}
		return
	}
	response.Success(c, "")
}

// DeleteDocumentSharePermission 移除分享权限
func DeleteDocumentSharePermission(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	permissionId := str.DefaultToInt(c.Query("share_id"), 0)
	if permissionId <= 0 {
		response.BadRequest(c, "参数错误：share_id")
		return
	}
	if err := services.NewDocumentService().DocumentPermissionService.HardDelete(
		"document_permission.resource_type = ? and document_permission.id", models.ResourceTypeDoc, permissionId,
		services.JoinArgs{Join: "inner join document on document.id = document_permission.resource_id", Args: nil},
		services.WhereArgs{Query: "document.user_id = ?", Args: []any{userId}},
	); err != nil {
		if err == services.ErrRecordNotFound {
			response.Unauthorized(c)
		} else {
			response.Fail(c, "删除错误")
		}
		return
	}
	response.Success(c, "")
}

type ApplyDocumentPermissionReq struct {
	DocId          string          `json:"doc_id" binding:"required"`
	PermType       models.PermType `json:"perm_type"`
	ApplicantNotes string          `json:"applicant_notes" binding:"required"`
}

// ApplyDocumentPermission 申请文档权限
func ApplyDocumentPermission(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req ApplyDocumentPermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(req.DocId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if err := documentService.GetById(documentId, &document); err != nil {
		if err == services.ErrRecordNotFound {
			response.BadRequest(c, "文档不存在")
		} else {
			response.Fail(c, "查询错误")
		}
		return
	}
	if document.UserId == userId {
		response.BadRequest(c, "不能申请自己的文档")
		return
	}
	if document.DocType != models.DocTypeShareable {
		response.Unauthorized(c)
		return
	}
	permType := req.PermType
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	documentInfoQueryRes := documentService.GetDocumentInfoByDocumentAndUserId(documentId, userId, permType)
	if permType <= documentInfoQueryRes.DocumentPermission.PermType {
		response.BadRequest(c, "已拥有权限")
		return
	}
	if documentInfoQueryRes.ApplicationCount >= 3 {
		response.BadRequest(c, "申请次数已达上限")
		return
	}
	if documentInfoQueryRes.SharesCount >= 5 && documentInfoQueryRes.DocumentPermission.PermType == 0 {
		response.BadRequest(c, "分享数量已达上限")
		return
	}
	if err := documentService.DocumentPermissionRequestsService.Create(&models.DocumentPermissionRequests{
		UserId:         userId,
		DocumentId:     documentId,
		PermType:       permType,
		ApplicantNotes: req.ApplicantNotes,
	}); err != nil {
		response.Fail(c, "新建错误")
		return
	}
	response.Success(c, "")
}

// GetDocumentPermissionRequestsList 获取申请列表
func GetDocumentPermissionRequestsList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := str.DefaultToInt(c.Query("doc_id"), 0)
	if documentId <= 0 {
		documentId = 0
	}
	var startTime *myTime.Time
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		t := myTime.Time(time.UnixMilli(startTimeInt))
		startTime = &t
	}
	response.Success(c, services.NewDocumentService().FindPermissionRequests(userId, documentId, startTime.String()))
}

// ReviewDocumentPermissionRequest 权限申请审核
func ReviewDocumentPermissionRequest(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentPermissionRequestsId := str.DefaultToInt(c.Query("apply_id"), 0)
	if documentPermissionRequestsId <= 0 {
		response.BadRequest(c, "参数错误：apply_id")
		return
	}
	approvalCode := str.DefaultToInt(c.Query("approval_code"), 0)
	if approvalCode < 0 || approvalCode > 1 {
		response.BadRequest(c, "参数错误：approval_code")
		return
	}
	documentService := services.NewDocumentService()
	var documentPermissionRequest models.DocumentPermissionRequests
	if err := documentService.DocumentPermissionRequestsService.Get(
		&documentPermissionRequest,
		"document_permission_requests.id = ? and document_permission_requests.status = ? and document.user_id = ?",
		documentPermissionRequestsId, models.StatusTypePending, userId,
		services.JoinArgs{Join: "inner join document on document.id = document_permission_requests.document_id", Args: nil},
	); err != nil {
		if err == services.ErrRecordNotFound {
			response.Unauthorized(c)
		} else {
			response.Fail(c, "查询错误")
		}
		return
	}
	if approvalCode == 0 {
		documentPermissionRequest.Status = models.StatusTypeDenied
	} else if approvalCode == 1 {
		documentPermissionRequest.Status = models.StatusTypeApproved
	}
	if err := documentService.DocumentPermissionRequestsService.UpdatesById(documentPermissionRequestsId, &documentPermissionRequest); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		var permType models.PermType
		_ = documentService.GetSelfDocumentPermissionByDocumentAndUserId(
			&permType,
			documentPermissionRequest.DocumentId,
			documentPermissionRequest.UserId,
		)
		if documentPermissionRequest.PermType <= permType {
			response.Success(c, "")
			return
		}
		if err := documentService.DocumentPermissionService.Create(&models.DocumentPermission{
			ResourceType: models.ResourceTypeDoc,
			ResourceId:   documentPermissionRequest.DocumentId,
			GranteeType:  models.GranteeTypeExternal,
			GranteeId:    documentPermissionRequest.UserId,
			PermType:     documentPermissionRequest.PermType,
		}); err != nil {
			response.Fail(c, "新建错误")
			return
		}
	}
	response.Success(c, "")
}

type UserDocumentPermResp struct {
	PermType models.PermType `json:"perm_type"`
}

// GetUserDocumentPerm 获取文档权限
func GetUserDocumentPerm(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := str.DefaultToInt(c.Query("doc_id"), 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	result := UserDocumentPermResp{}
	if err := services.NewDocumentService().GetDocumentPermissionByDocumentAndUserId(&result.PermType, documentId, userId); err != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	response.Success(c, result)
}
