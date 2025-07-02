package mongo_data

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
)

func TidyComments(removedDocuments []string, userIds []string) []models.UserComment {
	targetMongo := services.GetMongoDB()
	// 不存在的文档与用户的评论可以删除
	collection := targetMongo.DB.Collection("comment")
	ctx := context.Background()

	var deletedComments []models.UserComment

	// 1. 查找并删除不存在的用户的评论
	if len(userIds) > 0 {
		// 查找不在userIds列表中的用户评论
		userFilter := bson.M{
			"user": bson.M{"$nin": userIds},
		}

		// 先查询要删除的评论
		cursor, err := collection.Find(ctx, userFilter)
		if err != nil {
			log.Printf("查询用户评论失败: %v", err)
		} else {
			var userComments []models.UserComment
			if err := cursor.All(ctx, &userComments); err != nil {
				log.Printf("解析用户评论失败: %v", err)
			} else {
				deletedComments = append(deletedComments, userComments...)
				log.Printf("找到 %d 条不存在用户的评论", len(userComments))
			}
			cursor.Close(ctx)
		}

		// 删除不在userIds列表中的用户评论
		userDeleteResult, err := collection.DeleteMany(ctx, userFilter)
		if err != nil {
			log.Printf("删除用户评论失败: %v", err)
		} else {
			log.Printf("删除了 %d 条不存在用户的评论", userDeleteResult.DeletedCount)
		}
	}

	// 2. 查找并删除被移除文档的评论
	if len(removedDocuments) > 0 {
		// 查找属于removedDocuments的评论
		docFilter := bson.M{
			"document_id": bson.M{"$in": removedDocuments},
		}

		// 先查询要删除的评论
		cursor, err := collection.Find(ctx, docFilter)
		if err != nil {
			log.Printf("查询文档评论失败: %v", err)
		} else {
			var docComments []models.UserComment
			if err := cursor.All(ctx, &docComments); err != nil {
				log.Printf("解析文档评论失败: %v", err)
			} else {
				deletedComments = append(deletedComments, docComments...)
				log.Printf("找到 %d 条被删除文档的评论", len(docComments))
			}
			cursor.Close(ctx)
		}

		// 删除属于removedDocuments的评论
		docDeleteResult, err := collection.DeleteMany(ctx, docFilter)
		if err != nil {
			log.Printf("删除文档评论失败: %v", err)
		} else {
			log.Printf("删除了 %d 条被删除文档的评论", docDeleteResult.DeletedCount)
		}
	}

	log.Printf("总共删除了 %d 条评论", len(deletedComments))
	return deletedComments
}
