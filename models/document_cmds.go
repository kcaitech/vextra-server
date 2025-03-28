package models

import (
	"context"

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
	UnionId struct { // 联合id
		DocumentId int64  `json:"document_id" bson:"document_id"`
		CmdId      string `json:"cmd_id" bson:"cmd_id"`
	} `json:"union_id" bson:"_id"`
	DocumentId   int64  `json:"document_id" bson:"document_id"`
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
}

func NewCmdService(mongoDB *mongo.MongoDB) *CmdService {
	return &CmdService{
		MongoDB: mongoDB,
	}
}

// 获取DocumentId的文档中VerId范围从start到end的CmdItem
func (s *CmdService) GetCmdItems(documentId int64, verStart uint, verEnd uint) ([]CmdItem, error) {
	filter := bson.M{"document_id": documentId, "ver_id": bson.M{"$gte": verStart, "$lte": verEnd}}
	options := options.Find()
	options.SetSort(bson.D{{Key: "ver_id", Value: 1}})
	cursor, err := s.MongoDB.DB.Collection("document").Find(context.Background(), filter, options)
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
func (s *CmdService) GetCmdItemsFromStart(documentId int64, verStart uint) ([]CmdItem, error) {
	filter := bson.M{"document_id": documentId, "ver_id": bson.M{"$gte": verStart}}
	options := options.Find()
	options.SetSort(bson.D{{Key: "ver_id", Value: 1}})
	cursor, err := s.MongoDB.DB.Collection("cmd_items").Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	var cmdItems []CmdItem
	if err := cursor.All(context.Background(), &cmdItems); err != nil {
		return nil, err
	}
	return cmdItems, nil
}

func (s *CmdService) SaveCmdItems(cmdItems []CmdItem) (*mongodb.InsertManyResult, error) {
	// 设置联合id
	for i := range cmdItems {
		cmdItems[i].UnionId.DocumentId = cmdItems[i].DocumentId
		cmdItems[i].UnionId.CmdId = cmdItems[i].Cmd.Id
	}
	collection := s.MongoDB.DB.Collection("document")
	return collection.InsertMany(context.Background(), sliceutil.ConvertToAnySlice(cmdItems))
	// return err
}

// 获取DocumentId的文档中按VerId排序的最后一个CmdItem
func (s *CmdService) GetLastCmdItem(documentId int64) (*CmdItem, error) {

	filter := bson.M{"document_id": documentId}
	options := options.FindOne()
	options.SetSort(bson.D{{Key: "ver_id", Value: -1}})

	var cmdItem CmdItem
	err := s.MongoDB.DB.Collection("document").FindOne(context.Background(), filter, options).Decode(&cmdItem)
	if err == nil {
		return &cmdItem, nil
	}
	if err == mongodb.ErrNoDocuments {
		return nil, nil
	}
	return nil, err
}

// 获取特定id的cmd
