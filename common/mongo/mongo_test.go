package mongo

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"testing"
)

func Test0(t *testing.T) {
	if err := Init("config_test.yaml"); err != nil {
		log.Fatalln("mongo初始化失败：" + err.Error())
	}
	collection := DB.Collection("document")

	// 插入一个文档
	res, err := collection.InsertOne(nil, bson.M{"name": "John", "age": 30})
	if err != nil {
		log.Fatal(err)
	}
	id := res.InsertedID
	fmt.Println("Inserted document with ID:", id)

	// 查询插入的文档
	var result []struct {
		Name string
		Age  int
	}
	cur, err := collection.Find(nil, bson.M{"name": "John"})
	if err != nil {
		log.Fatal(err)
	}
	if err := cur.All(nil, &result); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Found a document:", result)
}
