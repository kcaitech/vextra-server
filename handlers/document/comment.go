package document

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	safereviewBase "kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	myTime "kcaitech.com/kcserver/utils/time"
)

func GetDocumentComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	documentId := c.Query("doc_id")
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType <= models.PermTypeNone {
		common.Forbidden(c, "")
		return
	}

	reqParams := bson.M{
		"document_id": documentId,
	}

	commentSrv := services.GetUserCommentService()
	documentCommentList, err := commentSrv.FindComments(reqParams)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	// 获取用户信息
	userIds := make([]string, 0)
	for _, comment := range documentCommentList {
		userIds = append(userIds, comment.User)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	result := make([]models.UserCommentWithUserInfo, 0)
	for _, comment := range documentCommentList {
		userId := comment.User
		userInfo, exists := userMap[userId]

		if exists {
			user := userInfo
			commentWithUser := models.UserCommentWithUserInfo{
				Id:                comment.Id,
				UserCommentCommon: comment.UserCommentCommon,
				User: models.UserProfile{
					Nickname: user.Nickname,
					Id:       user.UserID,
					Avatar:   user.Avatar,
				},
				CreatedAt:       comment.CreatedAt,
				RecordCreatedAt: comment.RecordCreatedAt,
			}
			result = append(result, commentWithUser)
		}
	}

	common.Success(c, &result)
}

func PostUserComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	var userComment models.UserCommentCommon
	if err := c.ShouldBindJSON(&userComment); err != nil {
		common.BadRequest(c, "")
		return
	}
	documentId := (userComment.DocumentId)
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	var permType models.PermType

	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeCommentable {
		common.Forbidden(c, "")
		return
	}

	userInfo, err := GetUserInfo(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	_userComment := models.UserComment{
		UserCommentCommon: models.UserCommentCommon{
			CommentId:  userComment.CommentId,
			ParentId:   userComment.ParentId,
			DocumentId: userComment.DocumentId,
			PageId:     userComment.PageId,
			ShapeId:    userComment.ShapeId,
			Content:    userComment.Content,
			OffsetX:    userComment.OffsetX,
			OffsetY:    userComment.OffsetY,
			RootX:      userComment.RootX,
			RootY:      userComment.RootY,
			Status:     models.UserCommentStatusCreated,
		},
		User:      userInfo.UserID,
		CreatedAt: myTime.Time(time.Now()).String(),
	}

	if _, err := myTime.Parse(_userComment.CreatedAt); err != nil {
		_userComment.RecordCreatedAt = _userComment.CreatedAt
	}

	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(userComment.Content)
		if err != nil {
			log.Println("评论审核失败", userComment.Content, err)
		} else if reviewResponse != nil && reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("评论审核不通过", userComment.Content, reviewResponse)
			var LockedWords string
			if wordsBytes, err := json.Marshal(reviewResponse.Words); err == nil {
				LockedWords = string(wordsBytes)
			}
			services.NewDocumentService().AddLocked(&models.DocumentLock{
				DocumentId:   documentId,
				LockedType:   models.LockedTypeComment,
				LockedReason: reviewResponse.Reason,
				LockedWords:  LockedWords,
				LockedTarget: userComment.CommentId,
			})
		}
	}

	commentSrv := services.GetUserCommentService()

	if err := commentSrv.InsertOne(&_userComment); err != nil {
		log.Println("mongo插入失败", err)
		common.ServerError(c, "评论失败")
		return
	}

	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type:    models.UserCommentPublishTypeAdd,
		Comment: _userComment.UserCommentCommon,
		User: models.UserProfile{
			Nickname: userInfo.Nickname,
			Id:       userInfo.UserID,
			Avatar:   userInfo.Avatar,
		},
		CreateAt: _userComment.CreatedAt,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+(documentId)+"]", publishData)
	}
	common.Success(c, _userComment.UserCommentCommon)
}

var errNoPermission = errors.New("无权限")

func checkUserPermission(userId string, commentId string, documentId string, expectPermType models.PermType) (*models.UserComment, error) {
	commentSrv := services.GetUserCommentService()
	comment, err := commentSrv.GetComment(documentId, commentId)
	if err != nil {
		return nil, err
	}

	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < expectPermType {
		return nil, errNoPermission
	}
	return comment, nil
}

func PutUserComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var userComment models.UserCommentCommon
	if err := c.ShouldBindJSON(&userComment); err != nil {
		log.Println("更新评论失败", err)
		common.BadRequest(c, "")
		return
	}
	documentId := (userComment.DocumentId)
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	// id是uuid
	// if str.DefaultToInt(userComment.Id, 0) <= 0 {
	// 	common.BadRequest(c, "参数错误：id")
	// 	return
	// }
	comment, err := checkUserPermission(userId, userComment.CommentId, documentId, models.PermTypeCommentable)
	if err != nil {
		if errors.Is(err, errNoPermission) {
			common.Forbidden(c, "")
			return
		} else {
			common.BadRequest(c, err.Error())
			return
		}
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(documentId, &document) != nil {
		common.BadRequest(c, "文档不存在")
		return
	}
	if comment.User != (userId) && document.UserId != userId {
		common.Forbidden(c, "")
		return
	}

	reviewClient := services.GetSafereviewClient()
	if userComment.Content != "" && reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(userComment.Content)
		if err != nil {
			log.Println("评论审核失败", userComment.Content, err)
			common.BadRequest(c, "审核失败")
			return
		} else if reviewResponse != nil && reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("评论审核不通过", userComment.Content, reviewResponse)
			common.BadRequest(c, "审核不通过")
			return
		}
	}

	commentSrv := services.GetUserCommentService()
	err = commentSrv.Update(comment, &userComment)
	if err != nil {
		log.Println("mongo更新失败", err)
		common.ServerError(c, "更新失败")
	}
	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type:    models.UserCommentPublishTypeUpdate,
		Comment: userComment,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+(comment.DocumentId)+"]", publishData)
	}
	common.Success(c, &userComment)
}

func DeleteUserComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	commentId := c.Query("comment_id")
	if commentId == "" {
		common.BadRequest(c, "参数错误：comment_id")
		return
	}
	documentId := c.Query("doc_id")
	// documentIdInt := str.DefaultToInt(documentId, 0)
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	// if str.DefaultToInt(commentId, 0) <= 0 {
	// 	common.BadRequest(c, "参数错误：comment_id")
	// 	return
	// }
	comment, err := checkUserPermission(userId, commentId, documentId, models.PermTypeCommentable)
	if err != nil {
		if errors.Is(err, errNoPermission) {
			common.Forbidden(c, "")
			return
		} else {
			common.BadRequest(c, err.Error())
			return
		}
	}
	var document models.Document
	if services.NewDocumentService().GetById(documentId, &document) != nil {
		common.BadRequest(c, "文档不存在")
		return
	}
	commentSrv := services.GetUserCommentService()
	if document.UserId != userId && comment.User != (userId) {
		if comment.ParentId == "" {
			common.Forbidden(c, "")
			return
		}
		comment, err := commentSrv.GetComment(documentId, comment.ParentId)
		if err != nil {
			fmt.Println("文档数据错误1", err)
			common.ServerError(c, "文档数据错误")
			return
		}
		if comment.User != (userId) {
			common.Forbidden(c, "")
			return
		}
	}
	delres, err := commentSrv.DeleteOne(comment)
	if err != nil || delres.DeletedCount <= 0 {
		log.Println("mongo删除失败", err)
		common.ServerError(c, "删除失败")
		return
	}
	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type: models.UserCommentPublishTypeDel,
		Comment: models.UserCommentCommon{
			CommentId: commentId,
			ParentId:  comment.ParentId,
		},
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+(comment.DocumentId)+"]", publishData)
	}
	common.Success(c, gin.H{
		"deleted": delres.DeletedCount,
	})
}

func SetUserCommentStatus(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var userComment models.UserCommentSetStatus
	if err := c.ShouldBindJSON(&userComment); err != nil {
		common.BadRequest(c, "")
		return
	}
	// if str.DefaultToInt(userComment.Id, 0) <= 0 {
	// 	common.BadRequest(c, "参数错误：id")
	// 	return
	// }
	documentId := (userComment.DocumentId)
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	if userComment.Status < models.UserCommentStatusCreated || userComment.Status > models.UserCommentStatusResolved {
		common.BadRequest(c, "参数错误：status")
		return
	}
	comment, err := checkUserPermission(userId, userComment.Id, documentId, models.PermTypeCommentable)
	if err != nil {
		if errors.Is(err, errNoPermission) {
			common.Forbidden(c, "")
			return
		} else {
			common.BadRequest(c, err.Error())
			return
		}
	}
	if comment.User != (userId) {
		var count int64
		if services.NewDocumentService().Count(&count, "id = ? and user_id = ?", comment.DocumentId, userId) != nil || count <= 0 {
			common.Forbidden(c, "")
			return
		}
	}
	commentSrv := services.GetUserCommentService()

	if err := commentSrv.Update(comment, &userComment); err != nil {
		log.Println("mongo更新失败", err)
		common.ServerError(c, "更新失败")
	}

	comment.Status = userComment.Status

	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type:    models.UserCommentPublishTypeUpdate,
		Comment: comment.UserCommentCommon,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+(comment.DocumentId)+"]", publishData)
	}
	common.Success(c, comment.UserCommentCommon)
}
