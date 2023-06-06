package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

type UserCommentStatus uint8

const (
	UserCommentStatusCreated UserCommentStatus = iota
	UserCommentStatusResolved
)

type UserComment struct {
	Id              string            `json:"id" bson:"_id"`
	ParentId        string            `json:"parent_id" bson:"parent_id"`
	RootId          string            `json:"root_id" bson:"root_id"`
	DocumentId      string            `json:"doc_id" bson:"document_id" binding:"required"`
	PageId          string            `json:"page_id" bson:"page_id" binding:"required"`
	ShapeId         string            `json:"shape_id" bson:"shape_id" binding:"required"`
	TargetShapeId   string            `json:"target_shape_id" bson:"target_shape_id" binding:"required"`
	ShapeFrame      map[string]any    `json:"shape_frame" bson:"shape_frame"`
	UserId          string            `json:"user_id" bson:"user_id"`
	CreatedAt       string            `json:"created_at" bson:"created_at"`
	RecordCreatedAt string            `json:"record_created_at" bson:"record_created_at"`
	Content         string            `json:"content" bson:"content" binding:"required"`
	Status          UserCommentStatus `json:"status" bson:"status"`
}

func GetDocumentComment(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	documentCommentList := make([]UserComment, 0)
	reqParams := bson.M{}
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
		reqParams["status"] = UserCommentStatus(str.DefaultToInt(c.Query("status"), 0))
	}
	commentCollection := mongo.DB.Collection("comment")
	if cur, err := commentCollection.Find(nil, reqParams); err == nil {
		_ = cur.All(nil, &documentCommentList)
	}
	response.Success(c, &documentCommentList)
}

func PostUserComment(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var userComment UserComment
	if err := c.ShouldBindJSON(&userComment); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(userComment.DocumentId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeCommentable {
		response.Forbidden(c, "")
		return
	}
	userComment.Id = str.IntToString(snowflake.NextId())
	userComment.UserId = str.IntToString(userId)
	userComment.CreatedAt = myTime.Time(time.Now()).String()
	if _, err := myTime.Parse(userComment.CreatedAt); err != nil {
		userComment.RecordCreatedAt = userComment.CreatedAt
	}
	userComment.Status = UserCommentStatusCreated
	commentCollection := mongo.DB.Collection("comment")
	if _, err := commentCollection.InsertOne(nil, userComment); err != nil {
		log.Println("mongo插入失败", err)
		response.Fail(c, "评论失败")
	}
	response.Success(c, &userComment)
}

var errNoPermission = errors.New("无权限")

func checkUserPermission(userId int64, commentId string) (*UserComment, error) {
	commentCollection := mongo.DB.Collection("comment")
	commentRes := commentCollection.FindOne(nil, bson.M{"_id": commentId})
	if commentRes.Err() != nil {
		return nil, errors.New("评论不存在")
	}
	var comment UserComment
	if err := commentRes.Decode(&comment); err != nil {
		fmt.Println("文档数据错误", err)
		return nil, errors.New("文档数据错误")
	}
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, str.DefaultToInt(comment.DocumentId, 0), userId); err != nil || permType < models.PermTypeCommentable {
		return nil, errNoPermission
	}
	return &comment, nil
}

type UserCommentUpdate struct {
	Id            string         `json:"id" bson:"_id"`
	ParentId      string         `json:"parent_id" bson:"parent_id"`
	RootId        string         `json:"root_id" bson:"root_id"`
	PageId        string         `json:"page_id" bson:"page_id,omitempty"`
	ShapeId       string         `json:"shape_id" bson:"shape_id,omitempty"`
	TargetShapeId string         `json:"target_shape_id" bson:"target_shape_id,omitempty"`
	ShapeFrame    map[string]any `json:"shape_frame" bson:"shape_frame,omitempty"`
	Content       string         `json:"content" bson:"content,omitempty"`
}

func PutUserComment(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var userComment UserCommentUpdate
	if err := c.ShouldBindJSON(&userComment); err != nil {
		response.BadRequest(c, "")
		return
	}
	if str.DefaultToInt(userComment.Id, 0) <= 0 {
		response.BadRequest(c, "参数错误：id")
		return
	}
	comment, err := checkUserPermission(userId, userComment.Id)
	if err != nil {
		if err == errNoPermission {
			response.Forbidden(c, "")
			return
		} else {
			response.BadRequest(c, err.Error())
			return
		}
	}
	if comment.UserId != str.IntToString(userId) {
		response.Forbidden(c, "")
		return
	}
	commentCollection := mongo.DB.Collection("comment")
	if _, err := commentCollection.UpdateByID(nil, userComment.Id, bson.M{"$set": &userComment}); err != nil {
		log.Println("mongo更新失败", err)
		response.Fail(c, "更新失败")
	}
	response.Success(c, &userComment)
}

func DeleteUserComment(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	commentId := c.Query("comment_id")
	if str.DefaultToInt(commentId, 0) <= 0 {
		response.BadRequest(c, "参数错误：comment_id")
		return
	}
	comment, err := checkUserPermission(userId, commentId)
	if err != nil {
		if err == errNoPermission {
			response.Forbidden(c, "")
			return
		} else {
			response.BadRequest(c, err.Error())
			return
		}
	}
	commentCollection := mongo.DB.Collection("comment")
	if comment.UserId != str.IntToString(userId) {
		if str.DefaultToInt(comment.ParentId, 0) <= 0 {
			response.Forbidden(c, "")
			return
		}
		commentRes := commentCollection.FindOne(nil, bson.M{"_id": comment.ParentId})
		if commentRes.Err() == nil {
			response.Forbidden(c, "")
			return
		}
		var comment UserComment
		if err := commentRes.Decode(&comment); err != nil {
			fmt.Println("文档数据错误1", err)
			response.Fail(c, "文档数据错误")
			return
		}
		if comment.UserId != str.IntToString(userId) {
			response.Forbidden(c, "")
			return
		}
	}
	if _, err := commentCollection.DeleteOne(nil, bson.M{"_id": commentId}); err != nil {
		log.Println("mongo删除失败", err)
		response.Fail(c, "删除失败")
	}
	response.Success(c, "")
}

type UserCommentSetStatus struct {
	Id     string            `json:"id" bson:"_id"`
	Status UserCommentStatus `json:"status" bson:"status"`
}

func SetUserCommentStatus(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var userComment UserCommentSetStatus
	if err := c.ShouldBindJSON(&userComment); err != nil {
		response.BadRequest(c, "")
		return
	}
	if str.DefaultToInt(userComment.Id, 0) <= 0 {
		response.BadRequest(c, "参数错误：id")
		return
	}
	if userComment.Status < UserCommentStatusCreated || userComment.Status > UserCommentStatusResolved {
		response.BadRequest(c, "参数错误：status")
		return
	}
	comment, err := checkUserPermission(userId, userComment.Id)
	if err != nil {
		if err == errNoPermission {
			response.Forbidden(c, "")
			return
		} else {
			response.BadRequest(c, err.Error())
			return
		}
	}
	if comment.Status != UserCommentStatusCreated {
		response.Fail(c, "当前状态不可修改")
		return
	}
	if comment.UserId != str.IntToString(userId) {
		var count int64
		if services.NewDocumentService().Count(&count, "id = ? and user_id = ?", comment.DocumentId, userId) != nil || count <= 0 {
			response.Forbidden(c, "")
			return
		}
	}
	commentCollection := mongo.DB.Collection("comment")
	if _, err := commentCollection.UpdateByID(nil, userComment.Id, bson.M{"$set": &userComment}); err != nil {
		log.Println("mongo更新失败", err)
		response.Fail(c, "更新失败")
	}
	response.Success(c, &userComment)
}
