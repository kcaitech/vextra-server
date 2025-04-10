package models

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	CommentId     string            `json:"id" bson:"comment_id,omitempty" binding:"required"` // 前端生成,uuid
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
	Id                primitive.ObjectID `json:"-" bson:"_id"`
	UserCommentCommon `json:",inline" bson:",inline"`
	// UnionId           struct {
	// 	DocumentId string `json:"document_id" bson:"document_id"`
	// 	CommentId  string `json:"comment_id" bson:"comment_id"`
	// } `json:"union_id" bson:"_id"`
	User            string `json:"user" bson:"user"`
	CreatedAt       string `json:"created_at" bson:"created_at"`
	RecordCreatedAt string `json:"record_created_at" bson:"record_created_at"`
}

type UserCommentWithUserInfo struct {
	Id                primitive.ObjectID `json:"-" bson:"_id"`
	UserCommentCommon `json:",inline" bson:",inline"`
	// UnionId           struct {
	// 	DocumentId string `json:"document_id" bson:"document_id"`
	// 	CommentId  string `json:"comment_id" bson:"comment_id"`
	// } `json:"union_id" bson:"_id"`
	User            UserProfile `json:"user" bson:"user"`
	CreatedAt       string      `json:"created_at" bson:"created_at"`
	RecordCreatedAt string      `json:"record_created_at" bson:"record_created_at"`
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
	MongoDB    *mongo.MongoDB
	Collection *mongodb.Collection
}

func NewUserCommentService(mongoDB *mongo.MongoDB) *UserCommentService {
	collection := mongoDB.DB.Collection("comment")
	// 创建索引
	// 创建多个索引
	indexModels := []mongodb.IndexModel{
		{
			Keys:    bson.D{{Key: "document_id", Value: 1}, {Key: "comment_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	// 批量创建索引
	_, err := collection.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			fmt.Println("索引已存在")
		} else {
			log.Fatalf("创建索引失败: %v", err)
		}
	}

	return &UserCommentService{MongoDB: mongoDB,
		Collection: collection}
}

func (s *UserCommentService) GetUserComment(documentId string) ([]UserComment, error) {

	filter := bson.M{
		"document_id": (documentId),
	}
	options := options.Find()
	options.SetSort(bson.D{{Key: "record_created_at", Value: -1}})
	comments := make([]UserComment, 0)
	cur, err := s.Collection.Find(context.Background(), filter, options)
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
	// for i := range commentItems {
	// 	commentItems[i].UnionId.DocumentId = commentItems[i].DocumentId
	// 	commentItems[i].UnionId.CommentId = commentItems[i].Id
	// }
	collection := s.Collection
	return collection.InsertMany(context.Background(), sliceutil.ConvertToAnySlice(commentItems))
}

func (s *UserCommentService) GetComment(documentId, commentId string) (*UserComment, error) {
	commentRes := s.Collection.FindOne(context.Background(), bson.M{
		"document_id": documentId,
		"comment_id":  commentId,
	})
	if commentRes.Err() != nil {
		return nil, commentRes.Err()
	}
	var comment UserComment
	if err := commentRes.Decode(&comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

func (s *UserCommentService) FindComments(filter bson.M) ([]UserComment, error) {

	options := options.Find()
	options.SetSort(bson.D{{Key: "record_created_at", Value: -1}, {Key: "_id", Value: -1}})
	cur, err := s.Collection.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}

	documentCommentList := make([]UserComment, 0)
	err = cur.All(context.Background(), &documentCommentList)
	if err != nil {
		return nil, err
	}
	return documentCommentList, nil
}

func (s *UserCommentService) InsertOne(_userComment *UserComment) error {
	_, err := s.Collection.InsertOne(context.Background(), _userComment)
	return err
}

func (s *UserCommentService) DeleteOne(_userComment *UserComment) (*mongodb.DeleteResult, error) {
	return s.Collection.DeleteOne(context.Background(), _userComment)
}

func (s *UserCommentService) Update(comment *UserComment, update interface{}) error {
	_, err := s.Collection.UpdateByID(context.Background(), comment.Id, bson.M{"$set": update})
	return err
}
