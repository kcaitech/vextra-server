package controllers

import (
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
)

type UserComment struct {
	Id              string            `json:"id" bson:"id"`
	ParentId        string            `json:"parent_id" bson:"parent_id"`
	RootId          string            `json:"root_id" bson:"root_id"`
	DocumentId      string            `json:"doc_id" bson:"document_id" binding:"required"`
	PageId          string            `json:"page_id" bson:"page_id" binding:"required"`
	ShapeId         string            `json:"shape_id" bson:"shape_id" binding:"required"`
	TargetShapeId   string            `json:"target_shape_id" bson:"target_shape_id" binding:"required"`
	ShapeFrame      map[string]any    `json:"shape_frame" bson:"shape_frame"`
	UserId          string            `json:"user_id" bson:"user_id"`
	CreatedAt       myTime.Time       `json:"created_at" bson:"created_at"`
	RecordCreatedAt myTime.Time       `json:"record_created_at" bson:"record_created_at"`
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
		reqParams["user_id"] = UserCommentStatus(str.DefaultToInt(c.Query("status"), 0))
	}
	commentCollection := mongo.DB.Collection("comment")
	if cur, err := commentCollection.Find(nil, reqParams); err == nil {
		_ = cur.All(nil, &documentCommentList)
	}
	response.Success(c, documentCommentList)
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
	userComment.CreatedAt = myTime.Time(time.Now())
	if userComment.RecordCreatedAt.IsZero() {
		userComment.RecordCreatedAt = userComment.CreatedAt
	}
	userComment.Status = UserCommentStatusCreated
	commentCollection := mongo.DB.Collection("comment")
	if _, err := commentCollection.InsertOne(nil, userComment); err != nil {
		log.Println("mongo插入失败", err)
		response.Fail(c, "评论失败")
	}
	response.Success(c, userComment)
}

func PutUserComment(c *gin.Context) {

}
