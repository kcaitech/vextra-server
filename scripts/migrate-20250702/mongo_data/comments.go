package mongo_data

import (
	"context"
	"log"
	"strconv"

	"github.com/google/uuid"
	"kcaitech.com/kcserver/providers/mongo"
)

func MigrateComments(sourceMongo *mongo.MongoDB, targetMongo *mongo.MongoDB, getUserID func(int64) (string, error)) {
	// 2. 迁移MongoDB数据 迁移评论数据
	log.Println("Migrating MongoDB data comments...")
	commentCollection := sourceMongo.DB.Collection("comment")
	commentCursor, err := commentCollection.Find(context.Background(), map[string]interface{}{})
	if err != nil {
		log.Fatalf("Error querying comments: %v", err)
	}
	defer commentCursor.Close(context.Background())

	var newComments []interface{}
	for commentCursor.Next(context.Background()) {
		var oldComment map[string]interface{}
		if err := commentCursor.Decode(&oldComment); err != nil {
			log.Printf("Error decoding comment: %v", err)
			continue
		}

		// 创建新格式的评论
		newComment := map[string]interface{}{}

		// 生成新的comment_id (UUID格式)
		commentId := uuid.New().String()

		// 基本字段转换
		newComment["parent_id"] = oldComment["parent_id"]
		newComment["document_id"] = oldComment["document_id"]
		newComment["page_id"] = oldComment["page_id"]
		newComment["shape_id"] = oldComment["target_shape_id"]
		newComment["content"] = oldComment["content"]
		newComment["status"] = oldComment["status"]
		newComment["created_at"] = oldComment["created_at"]
		newComment["record_created_at"] = oldComment["record_created_at"]
		newComment["comment_id"] = commentId

		// 提取用户ID
		if userObj, ok := oldComment["user"].(map[string]interface{}); ok {
			if userId, ok := userObj["id"].(string); ok {
				if oldId, err := strconv.ParseInt(userId, 10, 64); err == nil {
					userId, err := getUserID(oldId)
					if err != nil {
						log.Printf("Error getting user ID for comment %d: %v", oldComment["id"], err)
						continue
					}
					newComment["user"] = userId
				}
			}
		}

		// 转换位置信息
		if shapeFrame, ok := oldComment["shape_frame"].(map[string]interface{}); ok {
			x1, _ := shapeFrame["x1"].(float64)
			y1, _ := shapeFrame["y1"].(float64)
			x2, _ := shapeFrame["x2"].(float64)
			y2, _ := shapeFrame["y2"].(float64)

			newComment["offset_x"] = x2
			newComment["offset_y"] = y2
			newComment["root_x"] = x1
			newComment["root_y"] = y1
		}

		newComments = append(newComments, newComment)
	}
	if len(newComments) > 0 {
		log.Printf("Inserting %d comments", len(newComments))
		for _, comment := range newComments {
			commentMap := comment.(map[string]interface{})
			// 检查评论是否存在
			count, err := targetMongo.DB.Collection("comment").CountDocuments(context.Background(), map[string]interface{}{
				"document_id": commentMap["document_id"],
				"page_id":     commentMap["page_id"],
				"shape_id":    commentMap["shape_id"],
				"created_at":  commentMap["created_at"],
			}, nil)
			if err != nil {
				log.Printf("Error checking comment existence: %v", err)
				continue
			}
			if count > 0 {
				// 评论存在，执行更新
				if _, err := targetMongo.DB.Collection("comment").UpdateOne(context.Background(), map[string]interface{}{
					"document_id": commentMap["document_id"],
					"page_id":     commentMap["page_id"],
					"shape_id":    commentMap["shape_id"],
					"created_at":  commentMap["created_at"],
				}, map[string]interface{}{
					"$set": commentMap,
				}); err != nil {
					log.Printf("Error updating comment: %v", err)
				}
			} else {
				// 评论不存在，执行插入
				if _, err := targetMongo.DB.Collection("comment").InsertOne(context.Background(), comment); err != nil {
					log.Printf("Error inserting comment: %v", err)
				}
			}
		}
	}
}
