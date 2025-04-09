package models

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kcaitech.com/kcserver/providers/mongo"
	"kcaitech.com/kcserver/utils/sliceutil"
)

// for mongo
type Cmd struct {
	Id          string   `json:"id" bson:"cmd_id"`         // 前端过来的uuid，在文档内保证唯一就行
	BaseVer     uint     `json:"base_ver" bson:"base_ver"` // 引用的是VerId
	BatchId     string   `json:"batch_id" bson:"batch_id"` // 引用的是Id，uuid
	Ops         []bson.M `json:"ops" bson:"ops"`
	IsRecovery  bool     `json:"recovery" bson:"recovery"`
	Description string   `json:"description" bson:"description"`
	Time        int64    `json:"time" bson:"time"`                           // 编辑时间
	Posttime    int64    `json:"posttime" bson:"posttime"`                   // 上传时间
	DataFmtVer  string   `json:"fmt_ver,omitempty" bson:"fmt_ver,omitempty"` // int | string
}

// type CmdItemExtra struct {
// 	// CmdId       string `json:"cmd_id" bson:"cmd_id"`
// }

type CmdItem struct {
	// UnionId struct { // 联合id
	// 	DocumentId string `json:"document_id" bson:"document_id"`
	// 	CmdId      string `json:"cmd_id" bson:"cmd_id"`
	// } `json:"union_id" bson:"_id"`
	DocumentId   string `json:"document_id" bson:"document_id"`
	Cmd          Cmd    `json:",inline" bson:",inline"`
	UserId       string `json:"user_id" bson:"user_id"`
	VerId        uint   `json:"ver_id" bson:"ver_id"`           // version id
	BatchStartId uint   `json:"batch_start" bson:"batch_start"` // 引用的是VerId
	BatchLength  uint   `json:"batch_length" bson:"batch_length"`
}

func (cmdItem CmdItem) MarshalJSON() ([]byte, error) {
	return MarshalJSON(cmdItem)
}

func RenewCmdIds(cmdItems []CmdItem) {
	// 要处理batch_id
	batchIdMap := make(map[string]string)
	for i := range cmdItems {
		item := &cmdItems[i]
		newid := uuid.NewString()
		batchIdMap[item.Cmd.Id] = newid
		item.Cmd.Id = newid
		item.Cmd.BatchId = batchIdMap[item.Cmd.BatchId]
	}
}

func RenewCmdVerIds(cmdItems []CmdItem) {
	// 要处理batch_start_id
	if len(cmdItems) == 0 {
		return
	}
	offset := cmdItems[0].VerId
	for i := range cmdItems {
		cmdItems[i].VerId -= offset
		cmdItems[i].BatchStartId -= offset
	}
}

// service

type CmdService struct {
	MongoDB *mongo.MongoDB
	// Redis   *redis.RedisDB // 用于同步锁及消息发布
	Collection *mongodb.Collection
}

func NewCmdService(mongoDB *mongo.MongoDB) *CmdService {

	collection := mongoDB.DB.Collection("document")

	// 创建索引
	// 创建多个索引
	indexModels := []mongodb.IndexModel{
		{
			Keys:    bson.D{{Key: "document_id", Value: 1}, {Key: "cmd_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "document_id", Value: 1}, {Key: "ver_id", Value: 1}},
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

	// fmt.Printf("Created indexes %v\n", indexNames)
	return &CmdService{
		MongoDB:    mongoDB,
		Collection: collection,
	}
}

// 获取DocumentId的文档中VerId范围从start到end的CmdItem
func (s *CmdService) GetCmdItems(documentId string, verStart uint, verEnd uint) ([]CmdItem, error) {
	filter := bson.M{"document_id": documentId, "ver_id": bson.M{"$gte": verStart, "$lte": verEnd}}
	options := options.Find()
	options.SetSort(bson.D{{Key: "ver_id", Value: 1}})
	cursor, err := s.Collection.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	var cmdItems []CmdItem
	if err := cursor.All(context.Background(), &cmdItems); err != nil {
		return nil, err
	}
	return cmdItems, nil
}

// 获取DocumentId的文档中VerId范围从start开始的所有的CmdItem
func (s *CmdService) GetCmdItemsFromStart(documentId string, verStart uint) ([]CmdItem, error) {
	filter := bson.M{"document_id": documentId, "ver_id": bson.M{"$gte": verStart}}
	options := options.Find()
	options.SetSort(bson.D{{Key: "ver_id", Value: 1}})
	cursor, err := s.MongoDB.DB.Collection("cmd_items").Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	// 使用 make 创建非 nil 的空切片
	cmdItems := make([]CmdItem, 0)
	if err := cursor.All(context.Background(), &cmdItems); err != nil {
		return nil, err
	}
	return cmdItems, nil
}

func (s *CmdService) SaveCmdItems(cmdItems []CmdItem) (*mongodb.InsertManyResult, error) {
	// 设置联合id
	// for i := range cmdItems {
	// 	cmdItems[i].UnionId.DocumentId = cmdItems[i].DocumentId
	// 	cmdItems[i].UnionId.CmdId = cmdItems[i].Cmd.Id
	// }
	collection := s.Collection
	return collection.InsertMany(context.Background(), sliceutil.ConvertToAnySlice(cmdItems))
	// return err
}

// 获取DocumentId的文档中按VerId排序的最后一个CmdItem
func (s *CmdService) GetLastCmdItem(documentId string) (*CmdItem, error) {

	filter := bson.M{"document_id": documentId}
	options := options.FindOne()
	options.SetSort(bson.D{{Key: "ver_id", Value: -1}})

	var cmdItem CmdItem
	err := s.Collection.FindOne(context.Background(), filter, options).Decode(&cmdItem)
	if err == nil {
		return &cmdItem, nil
	}
	if err == mongodb.ErrNoDocuments {
		return nil, nil
	}
	return nil, err
}

// 获取特定id的cmd
func (s *CmdService) GetCmd(document_id, cmd_id string) (*CmdItem, error) {
	item := &CmdItem{}
	if err := s.Collection.FindOne(context.Background(), bson.M{"document_id": document_id, "cmd_id": cmd_id}).Decode(item); err != nil {
		return nil, err
	}
	return item, nil
}
