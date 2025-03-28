package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kcaitech.com/kcserver/providers/mongo"
	"kcaitech.com/kcserver/utils/sliceutil"
	"kcaitech.com/kcserver/utils/str"
)

type UserCommentStatus uint8

const (
	UserCommentStatusCreated UserCommentStatus = iota
	UserCommentStatusResolved
)

type UserComment struct {
	UnionId struct {
		DocumentId int64  `json:"document_id" bson:"document_id"`
		CommentId  string `json:"comment_id" bson:"comment_id"`
	} `json:"union_id" bson:"union_id"`
	Id              string            `json:"id" bson:"id,omitempty"` // 前端生成,uuid
	ParentId        string            `json:"parent_id" bson:"parent_id"`
	RootId          string            `json:"root_id" bson:"root_id"`
	DocumentId      int64             `json:"doc_id" bson:"document_id" binding:"required"`
	PageId          string            `json:"page_id" bson:"page_id" binding:"required"`
	ShapeId         string            `json:"shape_id" bson:"shape_id" binding:"required"`
	TargetShapeId   string            `json:"target_shape_id" bson:"target_shape_id" binding:"required"`
	ShapeFrame      map[string]any    `json:"shape_frame" bson:"shape_frame"`
	User            string            `json:"user" bson:"user"`
	CreatedAt       string            `json:"created_at" bson:"created_at"`
	RecordCreatedAt string            `json:"record_created_at" bson:"record_created_at"`
	Content         string            `json:"content" bson:"content" binding:"required"`
	Status          UserCommentStatus `json:"status" bson:"status"`
}

type UserCommentPublishType uint8

const (
	UserCommentPublishTypeAdd UserCommentPublishType = iota
	UserCommentPublishTypeDel
	UserCommentPublishTypeUpdate
)

type UserCommentPublishData struct {
	Type    UserCommentPublishType `json:"type"`
	Comment any                    `json:"comment"`
}

type UserCommentUpdate struct {
	Id            string         `json:"id" bson:"_id"`
	ParentId      string         `json:"parent_id,omitempty" bson:"parent_id"`
	RootId        string         `json:"root_id,omitempty" bson:"root_id"`
	PageId        string         `json:"page_id,omitempty" bson:"page_id,omitempty"`
	ShapeId       string         `json:"shape_id,omitempty" bson:"shape_id,omitempty"`
	TargetShapeId string         `json:"target_shape_id,omitempty" bson:"target_shape_id,omitempty"`
	ShapeFrame    map[string]any `json:"shape_frame,omitempty" bson:"shape_frame,omitempty"`
	Content       string         `json:"content,omitempty" bson:"content,omitempty"`
}

type UserCommentSetStatus struct {
	Id     string            `json:"id" bson:"_id"`
	Status UserCommentStatus `json:"status" bson:"status"`
}

type UserCommentService struct {
	MongoDB *mongo.MongoDB
}

func NewUserCommentService(mongoDB *mongo.MongoDB) *UserCommentService {
	return &UserCommentService{MongoDB: mongoDB}
}

func (s *UserCommentService) GetUserComment(documentId int64) ([]UserComment, error) {

	filter := bson.M{
		"document_id": str.IntToString(documentId),
	}
	options := options.Find()
	options.SetSort(bson.D{{Key: "record_created_at", Value: -1}})
	comments := make([]UserComment, 0)
	cur, err := s.MongoDB.DB.Collection("comment").Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}

	err = cur.All(context.Background(), &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// save comment items
func (s *UserCommentService) SaveCommentItems(commentItems []UserComment) (*mongodb.InsertManyResult, error) {
	if len(commentItems) == 0 {
		return nil, nil
	}
	// 设置联合id
	for i := range commentItems {
		commentItems[i].UnionId.DocumentId = commentItems[i].DocumentId
		commentItems[i].UnionId.CommentId = commentItems[i].Id
	}
	collection := s.MongoDB.DB.Collection("comment")
	return collection.InsertMany(context.Background(), sliceutil.ConvertToAnySlice(commentItems))
}
