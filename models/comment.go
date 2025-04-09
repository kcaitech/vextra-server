package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kcaitech.com/kcserver/providers/mongo"
	"kcaitech.com/kcserver/utils/sliceutil"
)

type UserCommentStatus uint8

const (
	UserCommentStatusCreated UserCommentStatus = iota
	UserCommentStatusResolved
)

type UserCommentCommon struct {
	Id            string            `json:"id" bson:"id,omitempty" binding:"required"` // 前端生成,uuid
	ParentId      string            `json:"parent_id" bson:"parent_id"`
	RootId        string            `json:"root_id" bson:"root_id"`
	DocumentId    string            `json:"doc_id,string" bson:"document_id" binding:"required"`
	PageId        string            `json:"page_id" bson:"page_id" binding:"required"`
	ShapeId       string            `json:"shape_id" bson:"shape_id" binding:"required"`
	TargetShapeId string            `json:"target_shape_id" bson:"target_shape_id" binding:"required"`
	ShapeFrame    map[string]any    `json:"shape_frame" bson:"shape_frame"`
	Content       string            `json:"content" bson:"content" binding:"required"`
	Status        UserCommentStatus `json:"status" bson:"status"`
}

// type UserCommentUpdate struct {
// 	Id            string         `json:"id" bson:"id,omitempty" binding:"required"` // 前端生成,uuid
// 	ParentId      string         `json:"parent_id" bson:"parent_id"`
// 	RootId        string         `json:"root_id" bson:"root_id"`
// 	DocumentId    string         `json:"doc_id" bson:"document_id" binding:"required"`
// 	PageId        string         `json:"page_id" bson:"page_id" binding:"required"`
// 	ShapeId       string         `json:"shape_id" bson:"shape_id" binding:"required"`
// 	TargetShapeId string         `json:"target_shape_id" bson:"target_shape_id" binding:"required"`
// 	ShapeFrame    map[string]any `json:"shape_frame" bson:"shape_frame"`
// 	Content       string         `json:"content" bson:"content" binding:"required"`
// }

type UserComment struct {
	UserCommentCommon `json:",inline" bson:",inline"`
	UnionId           struct {
		DocumentId string `json:"document_id" bson:"document_id"`
		CommentId  string `json:"comment_id" bson:"comment_id"`
	} `json:"union_id" bson:"_id"`
	User            string `json:"user" bson:"user"`
	CreatedAt       string `json:"created_at" bson:"created_at"`
	RecordCreatedAt string `json:"record_created_at" bson:"record_created_at"`
}

type UserCommentPublishType uint8

const (
	UserCommentPublishTypeAdd UserCommentPublishType = iota
	UserCommentPublishTypeDel
	UserCommentPublishTypeUpdate
)

type UserCommentPublishData struct {
	Type    UserCommentPublishType `json:"type"`
	Comment UserCommentCommon      `json:"comment"`
}

// type UserCommentUpdate struct {
// 	UserCommentCommon
// 	DocumentId int64 `json:"doc_id" bson:"document_id" binding:"required,string"`
// }

type UserCommentSetStatus struct {
	DocumentId string            `json:"doc_id" bson:"document_id" binding:"required"`
	Id         string            `json:"id" bson:"id" binding:"required"`
	Status     UserCommentStatus `json:"status" bson:"status"`
}

type UserCommentService struct {
	MongoDB *mongo.MongoDB
}

func NewUserCommentService(mongoDB *mongo.MongoDB) *UserCommentService {
	return &UserCommentService{MongoDB: mongoDB}
}

func (s *UserCommentService) GetUserComment(documentId string) ([]UserComment, error) {

	filter := bson.M{
		"document_id": (documentId),
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
