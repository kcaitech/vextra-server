package models

import (
	"fmt"
	"strings"

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
	DB, err := gorm.Open(mysql.Open(config.DB.DSN()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数形式的表名
		},
	})
	if err != nil {
		// log.Fatalf("连接数据库失败: %v", err)
		return nil, err
	}
	DB.Logger = DB.Logger.LogMode(logger.Silent)
	//DB = DB.Debug()
	//DB.Logger = DB.Logger.LogMode(logger.Info)

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
		return err
	}
	// document_favorites
	err = DocumentFavorites{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// document_permission_requests
	err = DocumentPermissionRequests{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// document_permission
	err = DocumentPermission{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// document_version
	err = DocumentVersion{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// document_lock
	err = DocumentLock{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// document
	err = Document{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// feedback
	err = Feedback{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// project
	err = Project{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// project_member
	err = ProjectMember{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// project_join_request
	err = ProjectJoinRequest{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// project_join_request_message_show
	err = ProjectJoinRequestMessageShow{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// project_favorite
	err = ProjectFavorite{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// team
	err = Team{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// team_member
	err = TeamMember{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// team_join_request
	err = TeamJoinRequest{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// team_join_request_message_show
	err = TeamJoinRequestMessageShow{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// user_kv_storage
	err = UserKVStorage{}.AutoMigrate(module.DB)
	if err != nil {
		return err
	}
	// 这两个不是这里实现的
	// user
	// user_profile
	return nil
}
