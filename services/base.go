package services

import (
	"log"

	"kcaitech.com/kcserver/config"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/auth"
	"kcaitech.com/kcserver/providers/mongo"
	"kcaitech.com/kcserver/providers/redis"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/providers/storage"
)

// DBModule 的单例
var dbModule *models.DBModule

func InitDBModule(config *models.DBModuleConfig) (*models.DBModule, error) {
	if dbModule != nil {
		return dbModule, nil
	}
	var err error
	dbModule, err = models.NewDBModule(config)
	if err != nil {
		return nil, err
	}
	err = dbModule.AutoMigrate()
	if err != nil {
		return nil, err
	}
	return dbModule, err
}

func GetDBModule() *models.DBModule {
	if dbModule == nil {
		panic("dbModule is nil")
	}
	return dbModule
}

// RedisDB 的单例
var redisDB *redis.RedisDB

func InitRedisDB(config *redis.RedisConf) (*redis.RedisDB, error) {
	if redisDB != nil {
		return redisDB, nil
	}
	var err error
	redisDB, err = redis.NewRedisDB(config)
	return redisDB, err
}

func GetRedisDB() *redis.RedisDB {
	if redisDB == nil {
		panic("redisDB is nil")
	}
	return redisDB
}

// MongoDB 的单例
var mongoDB *mongo.MongoDB

func InitMongoDB(config *mongo.MongoConf) (*mongo.MongoDB, error) {
	if mongoDB != nil {
		return mongoDB, nil
	}
	var err error
	mongoDB, err = mongo.NewMongoDB(config)
	return mongoDB, err
}

func GetMongoDB() *mongo.MongoDB {
	if mongoDB == nil {
		panic("mongoDB is nil")
	}
	return mongoDB
}

// jwt client 的单例
var jwtClient *auth.KCAuthClient

func InitKCAuthClient(authServerURL, clientID, clientSecret string) (*auth.KCAuthClient, error) {
	if jwtClient != nil {
		return jwtClient, nil
	}
	// var err error
	jwtClient = auth.NewAuthClient(authServerURL, clientID, clientSecret)
	return jwtClient, nil
}

func GetKCAuthClient() *auth.KCAuthClient {
	if jwtClient == nil {
		panic("jwtClient is nil")
	}
	return jwtClient
}

// safereview client 的单例
var safereviewClient safereview.Client

func InitSafereviewClient(config *safereview.SafeReviewConf) (safereview.Client, error) {
	if safereviewClient != nil {
		return safereviewClient, nil
	}
	var err error
	safereviewClient, err = safereview.NewSafeReview(config)
	if err != nil {
		return nil, err
	}
	return safereviewClient, nil
}

func GetSafereviewClient() safereview.Client {
	// if safereviewClient == nil {
	// 	panic("safereviewClient is nil")
	// }
	return safereviewClient
}

// storage client 的单例
var storageClient *storage.StorageClient

func InitStorageClient(config *storage.StorageConf) (*storage.StorageClient, error) {
	if storageClient != nil {
		return storageClient, nil
	}
	var err error
	storageClient, err = storage.NewStoraageClient(config)
	return storageClient, err
}

func GetStorageClient() *storage.StorageClient {
	return storageClient
}

var _config *config.Configuration

func GetConfig() *config.Configuration {
	return _config
}

var _cmdService *models.CmdService

func GetCmdService() *models.CmdService {
	return _cmdService
}

var _userCommentService *models.UserCommentService

func GetUserCommentService() *models.UserCommentService {
	return _userCommentService
}

// 初始化所有服务
func InitAllBaseServices(config *config.Configuration) error {
	_config = config
	var err error
	// 初始化数据库
	_, err = InitDBModule(&models.DBModuleConfig{
		DB: models.DBConfig{
			User:     config.DB.User,
			Password: config.DB.Password,
			Host:     config.DB.Host,
			Port:     config.DB.Port,
			Database: config.DB.Database,
		},
	})
	if err != nil {
		return err
	}
	// 初始化redis
	_, err = InitRedisDB(&config.Redis)
	if err != nil {
		return err
	}
	// 初始化mongo
	_, err = InitMongoDB(&config.Mongo)
	if err != nil {
		return err
	}
	// 初始化jwt
	_, err = InitKCAuthClient(config.AuthServerURL, config.AuthClientID, config.AuthClientSecret)
	if err != nil {
		return err
	}
	// 初始化safereview, 不是必须的
	_, err = InitSafereviewClient(&config.SafeReiew)
	if err != nil {
		// return err
		// 打印错误信息
		log.Printf("初始内容审核服务失败: %v", err)
	}
	// 初始化storage
	_, err = InitStorageClient(&config.Storage)
	if err != nil {
		return err
	}

	// 初始化cmdService
	_cmdService = models.NewCmdService(GetMongoDB())
	// 初始化userCommentService
	_userCommentService = models.NewUserCommentService(GetMongoDB())
	return nil
}
