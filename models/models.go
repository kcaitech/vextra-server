package models

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DBModule struct {
	DB *gorm.DB
}

type DBConfig struct {
	// DSN string `yaml:"url" json:"url"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Host     string `yaml:"host" json:"host"`
	Port     int64  `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
}

func (db *DBConfig) DSN() string {
	// root:kKEIjksvnOOIjdZ6rtzE@tcp(mysql:3306)/kcserver?charset=utf8&parseTime=True&loc=Local
	return db.User + ":" + db.Password + "@tcp(" + db.Host + ":" + fmt.Sprint(db.Port) + ")/" + db.Database + "?charset=utf8&parseTime=True&loc=Local"
}

type DBModuleConfig struct {
	DB         DBConfig `yaml:"db" json:"db"`
	StorageUrl struct {
		Document string `yaml:"document" json:"document"`
		Attatch  string `yaml:"attatch" json:"attatch"`
	} `yaml:"storage_url" json:"storage_url"`
}

func NewDBModule(config *DBModuleConfig) (*DBModule, error) {
	// GORM 配置
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数形式的表名
		},
		// 禁用默认事务
		SkipDefaultTransaction: true,
		// 准备语句
		PrepareStmt: true,
		// 日志配置
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond, // 慢 SQL 阈值
				LogLevel:                  logger.Silent,          // 日志级别
				IgnoreRecordNotFoundError: true,                   // 忽略记录未找到错误
				Colorful:                  false,                  // 禁用彩色打印
			},
		),
	}

	DB, err := gorm.Open(mysql.Open(config.DB.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取通用数据库对象 sql.DB，然后使用其提供的功能
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %v", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大生命周期

	return &DBModule{
		DB: DB,
	}, nil
}

func (module *DBModule) GetTableName(model any) string {
	db := module.DB.Model(model)
	_ = db.Statement.Parse(&model)
	return db.Statement.Table
}

func (module *DBModule) GetTableFieldNames(model any) []string {
	db := module.DB.Model(model)
	_ = db.Statement.Parse(&model)
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, field.DBName)
	}
	return fieldNames
}

func (module *DBModule) GetTableFieldNamesStr(model any) []string {
	db := module.DB.Model(model)
	_ = db.Statement.Parse(&model)
	tableName := db.Statement.Table
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s", tableName, field.DBName))
	}
	return fieldNames
}

func (module *DBModule) GetTableFieldNamesAliasByPrefix(model any, prefix string) []string {
	db := module.DB.Model(model)
	_ = db.Statement.Parse(&model)
	tableName := db.Statement.Table
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s as %s", tableName, field.DBName, prefix+field.DBName))
	}
	return fieldNames
}

func (module *DBModule) GetTableFieldNamesStrAliasByPrefix(model any, prefix string) string {
	return strings.Join(module.GetTableFieldNamesAliasByPrefix(model, prefix), ",")
}

func (module *DBModule) GetTableFieldNamesStrAliasByDefaultPrefix(model any, connector string) string {
	if connector == "" {
		connector = "__"
	}
	db := module.DB.Model(model)
	_ = db.Statement.Parse(&model)
	tableName := db.Statement.Table
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s as %s", tableName, field.DBName, tableName+connector+field.DBName))
	}
	return strings.Join(fieldNames, ",")
}

func (module *DBModule) AutoMigrate() error {
	var err error
	// document_access_record
	err = DocumentAccessRecord{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("DocumentAccessRecord:%s", err.Error())
	}
	// document_favorites
	err = DocumentFavorites{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("DocumentFavorites:%s", err.Error())
	}
	// document_permission_requests
	err = DocumentPermissionRequests{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("DocumentPermissionRequests:%s", err.Error())
	}
	// document_permission
	err = DocumentPermission{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("DocumentPermission:%s", err.Error())
	}
	// document_version
	err = DocumentVersion{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("DocumentVersion:%s", err.Error())
	}
	// document_lock
	err = DocumentLock{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("DocumentLock:%s", err.Error())
	}
	// document
	err = Document{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("Document:%s", err.Error())
	}
	// feedback
	err = Feedback{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("Feedback:%s", err.Error())
	}
	// project
	err = Project{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("Project:%s", err.Error())
	}
	// project_member
	err = ProjectMember{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("ProjectMember:%s", err.Error())
	}
	// project_join_request
	err = ProjectJoinRequest{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("ProjectJoinRequest:%s", err.Error())
	}
	// project_join_request_message_show
	err = ProjectJoinRequestMessageShow{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("ProjectJoinRequestMessageShow:%s", err.Error())
	}
	// project_favorite
	err = ProjectFavorite{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("ProjectFavorite:%s", err.Error())
	}
	// team
	err = Team{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("Team:%s", err.Error())
	}
	// team_member
	err = TeamMember{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("TeamMember:%s", err.Error())
	}
	// team_join_request
	err = TeamJoinRequest{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("TeamJoinRequest:%s", err.Error())
	}
	// team_join_request_message_show
	err = TeamJoinRequestMessageShow{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("TeamJoinRequestMessageShow:%s", err.Error())
	}
	// user_kv_storage
	err = UserKVStorage{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("UserKVStorage:%s", err.Error())
	}
	// resource_document
	err = ResourceDocument{}.AutoMigrate(module.DB)
	if err != nil {
		return fmt.Errorf("ResourceDocument:%s", err.Error())
	}

	// 这两个不是这里实现的
	// user
	// user_profile
	return nil
}
