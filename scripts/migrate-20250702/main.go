package main

import (
	"encoding/json"
	"fmt"

	"log"
	"os"

	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kcaitech.com/kcserver/scripts/migrate-20250702/storage"
	"kcaitech.com/kcserver/scripts/migrate-20250702/tables"
	"kcaitech.com/kcserver/services"

	config "kcaitech.com/kcserver/config"

	"kcaitech.com/kcserver/providers/mongo"

	migrate_config "kcaitech.com/kcserver/scripts/migrate-20250702/config"
	mongo_data "kcaitech.com/kcserver/scripts/migrate-20250702/mongo_data"
)

// NewWeixinUser 微信用户结构
type NewWeixinUser struct {
	UserID  string `json:"user_id" gorm:"primarykey"`
	UnionID string `json:"union_id" gorm:"unique"`
}

func main() {
	configDir := "" // 从当前目录加载
	conf, err := config.LoadYamlFile(configDir + "config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = services.InitAllBaseServices(conf)
	if err != nil {
		log.Fatalf("Error initializing services: %v", err)
	}
	log.Println("所有服务初始化成功")

	// 读取配置文件
	configFile, err := os.ReadFile(configDir + "migrate.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config migrate_config.Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	// 连接源数据库
	sourceDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Source.MySQL.User,
		config.Source.MySQL.Password,
		config.Source.MySQL.Host,
		config.Source.MySQL.Port,
		config.Source.MySQL.Database,
	)
	sourceDB, err := gorm.Open(mysql.Open(sourceDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Error connecting to source database: %v", err)
	}

	// 连接Auth数据库
	authDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Auth.MySQL.User,
		config.Auth.MySQL.Password,
		config.Auth.MySQL.Host,
		config.Auth.MySQL.Port,
		config.Auth.MySQL.Database,
	)
	authDB, err := gorm.Open(mysql.Open(authDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Error connecting to source auth database: %v", err)
	}

	var auth_users []NewWeixinUser
	if err := authDB.Table("weixin_users").Find(&auth_users).Error; err != nil {
		log.Fatalf("Error querying users: %v", err)
	}

	// 创建用户unionID映射
	wxUserIDMap := make(map[string]string)
	for _, user := range auth_users {
		wxUserIDMap[user.UnionID] = user.UserID
	}

	// 查询所有用户信息
	type User struct {
		ID        int64  `gorm:"column:id"`
		WxUnionID string `gorm:"column:wx_union_id"`
	}
	var users []User
	if err := sourceDB.Table("user").Find(&users).Error; err != nil {
		log.Fatalf("Error querying users: %v", err)
	}
	// 创建用户ID映射
	userIDMap := make(map[int64]string)
	for _, user := range users {
		if user.WxUnionID != "" {
			if wxUserID, ok := wxUserIDMap[user.WxUnionID]; ok {
				userIDMap[user.ID] = wxUserID
			} else {
				userIDMap[user.ID] = strconv.FormatInt(user.ID, 10)
			}
		} else {
			userIDMap[user.ID] = strconv.FormatInt(user.ID, 10)
		}
	}
	// 辅助函数：获取用户ID
	getUserID := func(oldUserID int64) (string, error) {
		if newID, ok := userIDMap[oldUserID]; ok {
			return newID, nil
		}
		return "", fmt.Errorf("user not found")
	}

	// 连接目标数据库
	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Target.MySQL.User,
		config.Target.MySQL.Password,
		config.Target.MySQL.Host,
		config.Target.MySQL.Port,
		config.Target.MySQL.Database,
	)
	targetDB, err := gorm.Open(mysql.Open(targetDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Error connecting to target database: %v", err)
	}

	// 连接源MongoDB
	sourceMongo, err := mongo.NewMongoDB(&mongo.MongoConf{
		Url: config.Source.Mongo.URL,
		Db:  config.Source.Mongo.DB,
	})
	if err != nil {
		log.Fatalf("Error connecting to source MongoDB: %v", err)
	}

	// 连接目标MongoDB
	targetMongo, err := mongo.NewMongoDB(&mongo.MongoConf{
		Url: config.Target.Mongo.URL,
		Db:  config.Target.Mongo.DB,
	})
	if err != nil {
		log.Fatalf("Error connecting to target MongoDB: %v", err)
	}

	// 开始迁移
	log.Println("Starting migration...")

	// 1. 迁移MySQL数据
	log.Println("Migrating MySQL data...")

	tables.MigrateDocument(sourceDB, targetDB, getUserID, config)
	tables.MigrateDocumentPermissionRequests(sourceDB, targetDB, getUserID)
	tables.MigrateDocumentVersion(sourceDB, targetDB)
	tables.MigrateDocumentPermission(sourceDB, targetDB)
	tables.MigrateDocumentAccessRecord(sourceDB, targetDB, getUserID)
	tables.MigrateDocumentFavorites(sourceDB, targetDB, getUserID)
	tables.MigrateTeam(sourceDB, targetDB)
	tables.MigrateTeamMember(sourceDB, targetDB, getUserID)
	tables.MigrateTeamJoinRequest(sourceDB, targetDB, getUserID)
	tables.MigrateTeamJoinRequestMessageShow(sourceDB, targetDB)
	tables.MigrateProject(sourceDB, targetDB)
	tables.MigrateProjectFavorite(sourceDB, targetDB, getUserID)
	tables.MigrateProjectJoinRequest(sourceDB, targetDB, getUserID)
	tables.MigrateProjectJoinRequestMessageShow(sourceDB, targetDB)
	tables.MigrateProjectMember(sourceDB, targetDB, getUserID)
	tables.MigrateFeedback(sourceDB, targetDB, getUserID)
	tables.MigrateUserKVStorage(sourceDB, targetDB, getUserID)

	mongo_data.MigrateComments(sourceMongo, targetMongo, getUserID)

	storage.MigrateTeamAvatars(config)

	log.Println("Migration completed!")
}
