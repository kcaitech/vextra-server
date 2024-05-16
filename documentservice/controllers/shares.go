package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/documentservice/config"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"strings"
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
	if _, err := services.NewDocumentService().DocumentPermissionService.HardDelete(
		"grantee_type = ? and grantee_id = ? and id = ?", models.GranteeTypeExternal, userId, permissionId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
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
	if err := documentService.Get(&document, "id = ?", documentId); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.Forbidden(c, "")
		} else {
			log.Println("查询错误", err)
			response.Fail(c, "")
		}
		return
	}
	if document.ProjectId == 0 && document.UserId != userId {
		response.Forbidden(c, "")
		return
	} else if document.ProjectId != 0 {
		projectService := services.NewProjectService()
		// 权限校验
		permType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		if err != nil || permType == nil {
			log.Println("权限查询错误", err)
			response.Fail(c, "")
			return
		}
		if (document.UserId != userId && *permType < models.ProjectPermTypeAdmin) || (document.UserId == userId && *permType < models.ProjectPermTypeCommentable) {
			response.Forbidden(c, "")
			return
		}
	}
	document.DocType = docType
	if _, err := documentService.UpdatesById(documentId, &document); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	if docType == models.DocTypePublicReadable || docType == models.DocTypePublicCommentable || docType == models.DocTypePublicEditable {
		var permType models.PermType
		switch docType {
		case models.DocTypePublicReadable:
			permType = models.PermTypeReadOnly
		case models.DocTypePublicCommentable:
			permType = models.PermTypeCommentable
		case models.DocTypePublicEditable:
			permType = models.PermTypeEditable
		}
		_, _ = documentService.DocumentPermissionService.UpdateColumns(map[string]any{
			"perm_type": permType,
		}, "resource_type = ? and resource_id = ? and grantee_type = ? and perm_source_type = ?", models.ResourceTypeDoc, documentId, models.GranteeTypeExternal, models.PermSourceTypeDefault)
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
	var document models.Document
	if err := documentService.Get(&document, "id = ?", documentId); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.Forbidden(c, "")
		} else {
			log.Println("查询错误", err)
			response.BadRequest(c, "文档不存在")
		}
		return
	}
	if document.ProjectId == 0 && document.UserId != userId {
		response.Forbidden(c, "")
		return
	} else if document.ProjectId != 0 {
		projectService := services.NewProjectService()
		permType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		if err != nil || permType == nil {
			log.Println("权限查询错误", err)
			response.Fail(c, "")
			return
		}
		if (document.UserId != userId && *permType < models.ProjectPermTypeAdmin) || (document.UserId == userId && *permType < models.ProjectPermTypeCommentable) {
			response.Forbidden(c, "")
			return
		}
	}
	response.Success(c, documentService.FindSharesByDocumentId(documentId))
}

type SetDocumentSharePermissionReq struct {
	ShareId  string          `json:"share_id" binding:"required"`
	PermType models.PermType `json:"perm_type"`
}

// SetDocumentSharePermission 修改分享权限
func SetDocumentSharePermission(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req SetDocumentSharePermissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	permissionId := str.DefaultToInt(req.ShareId, 0)
	if permissionId <= 0 {
		response.BadRequest(c, "参数错误：share_id")
		return
	}
	permType := req.PermType
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	documentPermissionService := services.NewDocumentService().DocumentPermissionService
	documentPermission, err := documentPermissionService.GetDocumentPermissionByPermId(permissionId)
	if err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			log.Println("数据不存在", err)
			response.BadRequest(c, "数据不存在")
		} else {
			log.Println("查询错误", err)
			response.Fail(c, "")
		}
		return
	}
	// 权限校验
	projectId := str.DefaultToInt(documentPermission.Document.ProjectId, 0)
	if projectId == 0 && documentPermission.Document.UserId != userId {
		response.Forbidden(c, "")
		return
	} else if projectId != 0 {
		projectService := services.NewProjectService()
		permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
		if err != nil || permType == nil {
			log.Println("权限查询错误", err)
			response.Fail(c, "")
			return
		}
		if (documentPermission.Document.UserId != userId && *permType < models.ProjectPermTypeAdmin) || (documentPermission.Document.UserId == userId && *permType < models.ProjectPermTypeCommentable) {
			response.Forbidden(c, "")
			return
		}
	}
	_, _ = services.NewDocumentService().DocumentPermissionService.UpdateColumns(
		map[string]any{"perm_type": permType, "perm_source_type": models.PermSourceTypeCustom},
		"id = ?",
		permissionId,
	)
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
	documentPermissionService := services.NewDocumentService().DocumentPermissionService
	var count int64
	if err := documentPermissionService.Count(
		&count,
		&services.JoinArgsRaw{Join: "inner join document on document.id = document_permission.resource_id", Args: nil},
		&services.WhereArgs{Query: "document_permission.resource_type = ? and document_permission.id = ? and document.user_id = ?", Args: []any{models.ResourceTypeDoc, permissionId, userId}},
	); err != nil || count <= 0 {
		if err != nil {
			response.Fail(c, "删除错误")
		} else {
			response.Forbidden(c, "")
		}
		return
	}
	_, _ = services.NewDocumentService().DocumentPermissionService.HardDelete("id = ?", permissionId)
	response.Success(c, "")
}

type ApplyDocumentPermissionReq struct {
	DocId          string          `json:"doc_id" binding:"required"`
	PermType       models.PermType `json:"perm_type"`
	ApplicantNotes string          `json:"applicant_notes"`
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
		if errors.Is(err, services.ErrRecordNotFound) {
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
		response.Forbidden(c, "")
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
	//if documentInfoQueryRes.SharesCount >= 5 && documentInfoQueryRes.DocumentPermission.PermType == 0 {
	//	response.BadRequest(c, "分享数量已达上限")
	//	return
	//}
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
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	documentService := services.NewDocumentService()
	result := documentService.FindPermissionRequests(userId, documentId, startTimeStr)
	if len(*result) > 0 {
		permissionRequestsIdList := sliceutil.MapT(func(item services.PermissionRequestsQueryResItem) int64 {
			return item.DocumentPermissionRequests.Id
		}, *result...)
		if _, err := documentService.DocumentPermissionRequestsService.UpdatesIgnoreZero(
			&models.DocumentPermissionRequests{FirstDisplayedAt: myTime.Time(time.Now())},
			"id in ?", permissionRequestsIdList,
		); err != nil {
			log.Println(err)
		}
	}
	response.Success(c, result)
}

type ReviewDocumentPermissionRequestReq struct {
	ApplyId      string `json:"apply_id" binding:"required"`
	ApprovalCode uint8  `json:"approval_code"`
}

// ReviewDocumentPermissionRequest 权限申请审核
func ReviewDocumentPermissionRequest(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req ReviewDocumentPermissionRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentPermissionRequestsId := str.DefaultToInt(req.ApplyId, 0)
	if documentPermissionRequestsId <= 0 {
		response.BadRequest(c, "参数错误：apply_id")
		return
	}
	approvalCode := req.ApprovalCode
	if approvalCode != 0 && approvalCode != 1 {
		response.BadRequest(c, "参数错误：approval_code")
		return
	}
	documentService := services.NewDocumentService()
	var documentPermissionRequest models.DocumentPermissionRequests
	if err := documentService.DocumentPermissionRequestsService.Get(
		&documentPermissionRequest,
		&services.JoinArgsRaw{Join: "inner join document on document.id = document_permission_requests.document_id", Args: nil},
		&services.WhereArgs{
			Query: "document_permission_requests.id = ? and document_permission_requests.status = ? and document.user_id = ?",
			Args:  []interface{}{documentPermissionRequestsId, models.StatusTypePending, userId},
		},
	); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "申请已被处理")
		} else {
			response.Fail(c, "查询错误")
		}
		return
	}
	if documentPermissionRequest.PermType < models.PermTypeReadOnly || documentPermissionRequest.PermType > models.PermTypeEditable {
		response.BadRequest(c, "参数错误：documentPermissionRequest.PermType")
		return
	}
	if approvalCode == 0 {
		documentPermissionRequest.Status = models.StatusTypeDenied
	} else if approvalCode == 1 {
		documentPermissionRequest.Status = models.StatusTypeApproved
	}
	documentPermissionRequest.ProcessedAt = myTime.Time(time.Now())
	documentPermissionRequest.ProcessedBy = userId
	if _, err := documentService.DocumentPermissionRequestsService.UpdatesById(documentPermissionRequestsId, &documentPermissionRequest); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		var permType models.PermType
		documentPermission, _, err := documentService.GetDocumentPermissionByDocumentAndUserId(
			&permType,
			documentPermissionRequest.DocumentId,
			documentPermissionRequest.UserId,
		)
		if err != nil {
			response.Fail(c, "查询错误")
			return
		}
		if documentPermission == nil {
			if err := documentService.DocumentPermissionService.Create(&models.DocumentPermission{
				ResourceType:   models.ResourceTypeDoc,
				ResourceId:     documentPermissionRequest.DocumentId,
				GranteeType:    models.GranteeTypeExternal,
				GranteeId:      documentPermissionRequest.UserId,
				PermType:       documentPermissionRequest.PermType,
				PermSourceType: models.PermSourceTypeCustom,
			}); err != nil {
				response.Fail(c, "新建错误")
				return
			}
		} else {
			if documentPermissionRequest.PermType <= permType {
				response.Success(c, "")
				return
			} else {
				documentPermission.PermType = documentPermissionRequest.PermType
				if _, err := documentService.DocumentPermissionService.UpdatesById(documentPermission.Id, documentPermission); err != nil {
					response.Fail(c, "更新错误")
					return
				}
			}
		}
	}
	response.Success(c, "")
}

type UserDocumentPermResp struct {
	PermType models.PermType `json:"perm_type"`
}

type wxMpAccessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
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
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&result.PermType, documentId, userId); err != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	response.Success(c, result)
}

// GetWxMpCode 获取微信小程序码
func GetWxMpCode(c *gin.Context) {
	scene := c.Query("scene")
	if len(scene) > 32 {
		response.BadRequest(c, "参数错误：scene")
		return
	}

	page := c.Query("page")

	// release trial develop
	envVersion := c.Query("env_version")
	if envVersion == "" {
		envVersion = "release"
	}
	if envVersion != "release" && envVersion != "trial" && envVersion != "develop" {
		response.BadRequest(c, "参数错误：env_version")
		return
	}

	// 发起请求
	// 获取AccessToken
	queryParams := url.Values{}
	queryParams.Set("appid", config.Config.WxMp.Appid)
	queryParams.Set("secret", config.Config.WxMp.Secret)
	queryParams.Set("grant_type", "client_credential")
	resp, err := http.Get("https://api.weixin.qq.com/cgi-bin/token" + "?" + queryParams.Encode())
	if err != nil {
		log.Println(err)
		response.Fail(c, "获取失败")
		return
	}
	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		response.Fail(c, "获取失败")
		return
	}
	// json解析
	var wxAccessTokenResp wxMpAccessTokenResp
	err = json.Unmarshal(body, &wxAccessTokenResp)
	if err != nil {
		log.Println(err)
		response.Fail(c, "获取失败")
		return
	}

	// 获取小程序码
	// 发起请求
	queryParams = url.Values{}
	queryParams.Set("access_token", wxAccessTokenResp.AccessToken)
	requestData := map[string]any{
		"check_path":  true,
		"env_version": envVersion,
		"width":       430,
		"auto_color":  false,
		"line_color": map[string]int{
			"r": 0,
			"g": 0,
			"b": 0,
		},
		"is_hyaline": false,
	}
	if scene != "" {
		// 前端使用了encodeURIComponent编码，这里要先解码
		scene, _ = url.QueryUnescape(scene)
		requestData["scene"] = scene
	}
	if page != "" {
		requestData["page"] = page
	}
	requestBodyBytes, _ := json.Marshal(requestData)
	resp, err = http.Post("https://api.weixin.qq.com/wxa/getwxacodeunlimit"+"?"+queryParams.Encode(), "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		log.Println(err)
		response.Fail(c, "获取失败")
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err = io.ReadAll(resp.Body)

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err != nil {
			log.Println(err)
			response.Fail(c, "获取失败")
			return
		}
		log.Println("获取小程序码失败", string(body))
		response.Fail(c, "获取失败")
		return
	}

	imageType := "image/png"
	if contentType != "" {
		imageType = contentType
	}
	response.Success(c, "data:"+imageType+";base64,"+base64.StdEncoding.EncodeToString(body))
}
