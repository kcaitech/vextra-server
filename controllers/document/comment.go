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
	"go.mongodb.org/mongo-driver/mongo/options"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	safereviewBase "kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
)

// type UserType struct {
// 	models.DefaultModelData
// 	Id       string `json:"id" bson:"id,omitempty"`
// 	Nickname string `json:"nickname" bson:"nickname,omitempty"`
// 	Avatar   string `json:"avatar" bson:"avatar,omitempty"`
// }

// func (user UserType) MarshalJSON() ([]byte, error) {
// 	if strings.HasPrefix(user.Avatar, "/") {
// 		user.Avatar = config.Config.StorageUrl.Attatch + user.Avatar
// 	}
// 	return models.MarshalJSON(user)
// }

func GetDocumentComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := c.Query("doc_id")
	documentIdInt := str.DefaultToInt(documentId, 0)
	if documentIdInt <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentIdInt, userId); err != nil || permType <= models.PermTypeNone {
		response.Forbidden(c, "")
		return
	}
	documentCommentList := make([]models.UserComment, 0)
	reqParams := bson.M{
		"document_id": documentId,
	}
	if pageId := c.Query("page_id"); pageId != "" {
		reqParams["page_id"] = pageId
	}
	if targetShapeId := c.Query("target_shape_id"); targetShapeId != "" {
		reqParams["target_shape_id"] = targetShapeId
	}
	if rootId := c.Query("root_id"); rootId != "" {
		reqParams["root_id"] = rootId
	}
	if parentId := c.Query("parent_id"); parentId != "" {
		reqParams["parent_id"] = parentId
	}
	if userId := c.Query("user_id"); userId != "" {
		reqParams["user_id"] = userId
	}
	if status := c.Query("status"); status != "" {
		reqParams["status"] = models.UserCommentStatus(str.DefaultToInt(c.Query("status"), 0))
	}
	mongoDB := services.GetMongoDB()
	commentCollection := mongoDB.DB.Collection("comment")
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"record_created_at", -1}, {"_id", -1}})
	if cur, err := commentCollection.Find(nil, reqParams, findOptions); err == nil {
		_ = cur.All(nil, &documentCommentList)
	}
	response.Success(c, &documentCommentList)
}

func PostUserComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var userComment models.UserComment
	if err := c.ShouldBindJSON(&userComment); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := userComment.DocumentId
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeCommentable {
		response.Forbidden(c, "")
		return
	}
	// 使用mongo的_id
	// userComment.Id = str.IntToString(snowflake.NextId())

	accessToken, _ := c.Get("access_token")
	if accessToken == nil {
		response.Unauthorized(c)
		return
	}
	jwtClient := services.GetJWTClient()
	userInfo, err := jwtClient.GetUserInfo(accessToken.(string))
	if err != nil {
		response.Unauthorized(c)
		return
	}
	// userComment.User = models.UserProfile{
	// 	UserId:   userInfo.UserID,
	// 	Nickname: userInfo.Profile.Nickname,
	// 	Avatar:   userInfo.Profile.Avatar,
	// }
	userComment.User = userInfo.UserID
	userComment.CreatedAt = myTime.Time(time.Now()).String()
	if _, err := myTime.Parse(userComment.CreatedAt); err != nil {
		userComment.RecordCreatedAt = userComment.CreatedAt
	}
	userComment.Status = models.UserCommentStatusCreated

	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(userComment.Content)
		if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("评论审核不通过", userComment.Content, err, reviewResponse)
			documentService := services.NewDocumentService()
			var document models.Document
			if documentService.GetById(documentId, &document) != nil {
				log.Println("文档不存在", documentId)
				response.Fail(c, "文档不存在")
				return
			}
			LockedAt := (time.Now())
			LockedReason := "文本审核不通过：" + reviewResponse.Reason
			var LockedWords string
			if wordsBytes, err := json.Marshal(reviewResponse.Words); err == nil {
				LockedWords = string(wordsBytes)
			}
			documentService.UpdateLocked(documentId, LockedAt, LockedReason, LockedWords)
		}
	}

	mongoDB := services.GetMongoDB()
	commentCollection := mongoDB.DB.Collection("comment")
	if _, err := commentCollection.InsertOne(nil, userComment); err != nil {
		log.Println("mongo插入失败", err)
		response.Fail(c, "评论失败")
	}
	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type:    models.UserCommentPublishTypeAdd,
		Comment: &userComment,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+str.IntToString(userComment.DocumentId)+"]", publishData)
	}
	response.Success(c, &userComment)
}

var errNoPermission = errors.New("无权限")

func checkUserPermission(userId string, commentId string, expectPermType models.PermType) (*models.UserComment, error) {
	mongoDB := services.GetMongoDB()
	commentCollection := mongoDB.DB.Collection("comment")
	commentRes := commentCollection.FindOne(nil, bson.M{"_id": commentId})
	if commentRes.Err() != nil {
		return nil, errors.New("评论不存在")
	}
	var comment models.UserComment
	if err := commentRes.Decode(&comment); err != nil {
		fmt.Println("评论数据错误", err)
		return nil, errors.New("评论数据错误")
	}
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, comment.DocumentId, userId); err != nil || permType < expectPermType {
		return nil, errNoPermission
	}
	return &comment, nil
}

func PutUserComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var userComment models.UserCommentUpdate
	if err := c.ShouldBindJSON(&userComment); err != nil {
		response.BadRequest(c, "")
		return
	}
	if str.DefaultToInt(userComment.Id, 0) <= 0 {
		response.BadRequest(c, "参数错误：id")
		return
	}
	comment, err := checkUserPermission(userId, userComment.Id, models.PermTypeCommentable)
	if err != nil {
		if errors.Is(err, errNoPermission) {
			response.Forbidden(c, "")
			return
		} else {
			response.BadRequest(c, err.Error())
			return
		}
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(comment.DocumentId, &document) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if comment.User != (userId) && document.UserId != userId {
		response.Forbidden(c, "")
		return
	}

	reviewClient := services.GetSafereviewClient()
	if userComment.Content != "" && reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(userComment.Content)
		if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("评论审核不通过", userComment.Content, err, reviewResponse)
			response.Fail(c, "审核不通过")
			return
		}
	}

	mongoDB := services.GetMongoDB()
	commentCollection := mongoDB.DB.Collection("comment")
	if _, err := commentCollection.UpdateByID(nil, userComment.Id, bson.M{"$set": &userComment}); err != nil {
		log.Println("mongo更新失败", err)
		response.Fail(c, "更新失败")
	}
	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type:    models.UserCommentPublishTypeUpdate,
		Comment: &userComment,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+str.IntToString(comment.DocumentId)+"]", publishData)
	}
	response.Success(c, &userComment)
}

func DeleteUserComment(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	commentId := c.Query("comment_id")
	if str.DefaultToInt(commentId, 0) <= 0 {
		response.BadRequest(c, "参数错误：comment_id")
		return
	}
	comment, err := checkUserPermission(userId, commentId, models.PermTypeCommentable)
	if err != nil {
		if errors.Is(err, errNoPermission) {
			response.Forbidden(c, "")
			return
		} else {
			response.BadRequest(c, err.Error())
			return
		}
	}
	var document models.Document
	if services.NewDocumentService().GetById(comment.DocumentId, &document) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	mongoDB := services.GetMongoDB()
	commentCollection := mongoDB.DB.Collection("comment")
	if document.UserId != userId && comment.User != (userId) {
		if str.DefaultToInt(comment.ParentId, 0) <= 0 {
			response.Forbidden(c, "")
			return
		}
		commentRes := commentCollection.FindOne(nil, bson.M{"_id": comment.ParentId})
		if commentRes.Err() == nil {
			response.Forbidden(c, "")
			return
		}
		var comment models.UserComment
		if err := commentRes.Decode(&comment); err != nil {
			fmt.Println("文档数据错误1", err)
			response.Fail(c, "文档数据错误")
			return
		}
		if comment.User != (userId) {
			response.Forbidden(c, "")
			return
		}
	}
	if _, err := commentCollection.DeleteOne(nil, bson.M{"_id": commentId}); err != nil {
		log.Println("mongo删除失败", err)
		response.Fail(c, "删除失败")
	}
	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type: models.UserCommentPublishTypeDel,
		Comment: &struct {
			Id string `json:"id" bson:"_id"`
		}{
			Id: commentId,
		},
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+str.IntToString(comment.DocumentId)+"]", publishData)
	}
	response.Success(c, "")
}

func SetUserCommentStatus(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var userComment models.UserCommentSetStatus
	if err := c.ShouldBindJSON(&userComment); err != nil {
		response.BadRequest(c, "")
		return
	}
	if str.DefaultToInt(userComment.Id, 0) <= 0 {
		response.BadRequest(c, "参数错误：id")
		return
	}
	if userComment.Status < models.UserCommentStatusCreated || userComment.Status > models.UserCommentStatusResolved {
		response.BadRequest(c, "参数错误：status")
		return
	}
	comment, err := checkUserPermission(userId, userComment.Id, models.PermTypeCommentable)
	if err != nil {
		if errors.Is(err, errNoPermission) {
			response.Forbidden(c, "")
			return
		} else {
			response.BadRequest(c, err.Error())
			return
		}
	}
	if comment.User != (userId) {
		var count int64
		if services.NewDocumentService().Count(&count, "id = ? and user_id = ?", comment.DocumentId, userId) != nil || count <= 0 {
			response.Forbidden(c, "")
			return
		}
	}
	mongoDB := services.GetMongoDB()
	commentCollection := mongoDB.DB.Collection("comment")
	if _, err := commentCollection.UpdateByID(nil, userComment.Id, bson.M{"$set": &userComment}); err != nil {
		log.Println("mongo更新失败", err)
		response.Fail(c, "更新失败")
	}
	if publishData, err := json.Marshal(&models.UserCommentPublishData{
		Type:    models.UserCommentPublishTypeUpdate,
		Comment: &userComment,
	}); err == nil {
		redisClient := services.GetRedisDB()
		redisClient.Client.Publish(context.Background(), "Document Comment[DocumentId:"+str.IntToString(comment.DocumentId)+"]", publishData)
	}
	response.Success(c, &userComment)
}
