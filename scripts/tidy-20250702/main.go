package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kcaitech.com/kcserver/config"
	"kcaitech.com/kcserver/services"

	tidy_config "kcaitech.com/kcserver/scripts/tidy-20250702/config"
	"kcaitech.com/kcserver/scripts/tidy-20250702/mongo_data"
	tidy_storage "kcaitech.com/kcserver/scripts/tidy-20250702/storage"
	"kcaitech.com/kcserver/scripts/tidy-20250702/tables"
)

type User struct { // Automatically generated ID
	UserID string `json:"user_id" gorm:"primarykey"` // Login identifier, for normal accounts this is the login account, for email accounts it's automatically generated
	// Password      string     `json:"-" gorm:"not null"`
	// Status        UserStatus `json:"status" gorm:"not null;default:'active'"`
	// Nickname      string     `json:"nickname" gorm:"size:50"` // Nickname
	// Avatar        string     `json:"avatar" gorm:"size:255"`  // Avatar URL
	// LastLogin     *time.Time `json:"last_login"`
	// LoginAttempts int        `json:"login_attempts" gorm:"default:0"`
	// LastAttempt   *time.Time `json:"last_attempt"`
	// CreatedAt     time.Time  `json:"created_at"`
	// UpdatedAt     time.Time  `json:"updated_at"`
}

func getAllUserIds(config *tidy_config.Config) []string {
	// 连接Auth数据库
	authDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config.AuthDB.User,
		config.AuthDB.Password,
		config.AuthDB.Host,
		config.AuthDB.Port,
		config.AuthDB.Database,
	)
	authDB, err := gorm.Open(mysql.Open(authDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Error connecting to source auth database: %v", err)
	}

	var auth_users []User
	if err := authDB.Table("users").Find(&auth_users).Error; err != nil {
		log.Fatalf("Error querying users: %v", err)
	}

	userIds := make([]string, len(auth_users))
	for i, user := range auth_users {
		userIds[i] = user.UserID
	}

	return userIds
}

func Init() *config.Configuration {

	configDir := "config/"
	conf, err := config.LoadYamlFile(configDir + "config.yaml")
	if err != nil {
		fmt.Println("err", err)
		panic(err)
	}
	fmt.Println("conf", conf)

	// 初始化services
	err = services.InitAllBaseServices(conf)
	if err != nil {
		log.Fatalf("kcserver服务初始化失败: %v", err)
	}

	return conf
}

func main() {
	Init()

	tidy_config, err := tidy_config.LoadYamlFile("config/tidy.yaml")
	if err != nil {
		log.Fatalf("Error loading tidy config: %v", err)
	}

	userIds := getAllUserIds(tidy_config)
	fmt.Println(userIds)

	db := services.GetDBModule().DB

	tables.TidyDocumentAccessRecord(db, userIds)
	tables.TidyDocumentFavorites(db, userIds)
	tables.TidyDocumentPermissionRequests(db, userIds)

	removedDocuments := tables.TidyDocument(db, userIds)

	removedDocumentIds := make([]string, len(removedDocuments))
	for i, document := range removedDocuments {
		removedDocumentIds[i] = document.Id
	}

	removedDocumentPaths := make([]string, len(removedDocuments))
	for i, document := range removedDocuments {
		removedDocumentPaths[i] = document.Path
	}

	tidy_storage.TidyDocumentStorage(removedDocumentPaths)
	tables.TidyDocumentPermission(db, removedDocumentIds)
	tables.TidyDocumentVersion(db, removedDocumentIds)
	tables.TidyFeedback(db, userIds)

	tables.TidyProjectFavorite(db, userIds)
	tables.TidyProjectJoinRequestMessageShow(db, userIds)
	tables.TidyProjectJoinRequest(db, userIds)
	tables.TidyProjectMember(db, userIds)

	tables.TidyTeamJoinRequestMessageShow(db, userIds)
	tables.TidyTeamJoinRequest(db, userIds)

	tables.TidyTeamMember(db, userIds)
	removedTeams := tables.TidyTeam(db)
	removedTeamIds := make([]string, len(removedTeams))
	for i, team := range removedTeams {
		removedTeamIds[i] = team.Id
	}

	removedTeamAvatars := make([]string, len(removedTeams))
	for i, team := range removedTeams {
		removedTeamAvatars[i] = team.Avatar
	}
	tidy_storage.TidyTeamAvatars(removedTeamAvatars)

	tables.TidyProject(db, removedTeamIds)
	tables.TidyUserKVStorage(db, userIds)

	mongo_data.TidyComments(removedDocumentIds, userIds)
}
