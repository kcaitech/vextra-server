package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"io"
	"log"
	"os"

	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kcaitech.com/kcserver/models"

	"net/http"

	config "kcaitech.com/kcserver/config"
	autoupdate "kcaitech.com/kcserver/handlers/document"

	"reflect"

	"kcaitech.com/kcserver/providers/mongo"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/str"
	utilTime "kcaitech.com/kcserver/utils/time"
)

type Config struct {
	Source struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		} `json:"mysql"`
		Mongo struct {
			URL string `json:"url"`
			DB  string `json:"db"`
		} `json:"mongo"`
		GenerateApiUrl string `json:"generateApiUrl"`
	} `json:"source"`
	Target struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		} `json:"mysql"`
		Mongo struct {
			URL string `json:"url"`
			DB  string `json:"db"`
		} `json:"mongo"`
	} `json:"target"`
	Auth struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		} `json:"mysql"`
	} `json:"auth"`
}

type NewWeixinUser struct {
	UserID  string `json:"user_id" gorm:"primarykey"`
	UnionID string `json:"union_id" gorm:"unique"`
}

func migrateDocumentStorage(documentId int64, generateApiUrl string) error {
	// 连接目标存储
	log.Println("documentId: ", documentId)

	// var generateApiUrl = "http://192.168.0.131:8088/generate" // 旧版本更新服务地址
	documentIdStr := str.IntToString(documentId)
	resp, err := http.Get(generateApiUrl + "?documentId=" + documentIdStr)
	if err != nil {
		log.Println(generateApiUrl, "http.NewRequest err", err)
		return err
	}

	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(generateApiUrl, "io.ReadAll err", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Println(generateApiUrl, "请求失败", resp.StatusCode, string(body))
		return errors.New("请求失败")
	}

	var version struct {
		LastCmdVerId string                `json:"lastCmdId"`
		DocumentData autoupdate.ExFromJson `json:"documentData"`
		DocumentText string                `json:"documentText"`
		MediasSize   uint64                `json:"mediasSize"`
		PageSvgs     []string              `json:"pageSvgs"`
	}
	err = json.Unmarshal(body, &version)
	if err != nil {
		log.Println(generateApiUrl, "resp", err)
		return err
	}

	log.Println("auto update document, start upload data", documentId)
	// upload document data
	header := autoupdate.Header{
		DocumentId:   documentIdStr,
		LastCmdVerId: version.LastCmdVerId,
	}
	response := autoupdate.Response{}
	data := autoupdate.UploadData{
		DocumentMeta: autoupdate.Data(version.DocumentData.DocumentMeta),
		Pages:        version.DocumentData.Pages,
		MediaNames:   version.DocumentData.MediaNames,
		MediasSize:   version.MediasSize,
		DocumentText: version.DocumentText,
		PageSvgs:     version.PageSvgs,
	}
	autoupdate.UploadDocumentData(&header, &data, nil, &response)
	if response.Status != autoupdate.ResponseStatusSuccess {
		log.Println("auto update failed", response.Message)
		return errors.New("auto update failed")
	}
	log.Println("auto update successed")
	return nil
}

func main() {
	configDir := "" // 从当前目录加载
	conf, err := config.LoadYamlFile(configDir + "config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	services.InitAllBaseServices(conf)

	// 读取配置文件
	configFile, err := os.ReadFile(configDir + "migrate.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
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
	log.Println("sourceDSN: ", sourceDSN)
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
	log.Println("authDSN: ", authDSN)
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
	getUserID := func(oldUserID int64) string {
		if newID, ok := userIDMap[oldUserID]; ok {
			return newID
		}
		return strconv.FormatInt(oldUserID, 10)
	}

	// 连接目标数据库
	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Target.MySQL.User,
		config.Target.MySQL.Password,
		config.Target.MySQL.Host,
		config.Target.MySQL.Port,
		config.Target.MySQL.Database,
	)
	log.Println("targetDSN: ", targetDSN)
	targetDB, err := gorm.Open(mysql.Open(targetDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	log.Println("targetDB: ", targetDB)
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
	log.Println("targetMongo: ", targetMongo, sourceMongo)

	// 开始迁移
	log.Println("Starting migration...")

	// 1. 迁移MySQL数据
	log.Println("Migrating MySQL data...")

	// 迁移文档表
	var oldDocuments []struct {
		ID           int64      `gorm:"column:id"`
		CreatedAt    time.Time  `gorm:"column:created_at"`
		UpdatedAt    time.Time  `gorm:"column:updated_at"`
		DeletedAt    *time.Time `gorm:"column:deleted_at"`
		UserId       int64      `gorm:"column:user_id"`
		Path         string     `gorm:"column:path"`
		DocType      uint8      `gorm:"column:doc_type"`
		Name         string     `gorm:"column:name"`
		Size         uint64     `gorm:"column:size"`
		PurgedAt     time.Time  `gorm:"column:purged_at"`
		DeleteBy     int64      `gorm:"column:delete_by"`
		VersionId    string     `gorm:"column:version_id"`
		TeamId       int64      `gorm:"column:team_id"`
		ProjectId    int64      `gorm:"column:project_id"`
		LockedAt     time.Time  `gorm:"column:locked_at"`
		LockedReason string     `gorm:"column:locked_reason"`
		LockedWords  string     `gorm:"column:locked_words"`
	}

	if err := sourceDB.Table("document").Order("created_at DESC").Find(&oldDocuments).Error; err != nil {
		log.Fatalf("Error querying documents: %v", err)
	}
	// var documentIds []int64
	for _, oldDoc := range oldDocuments {
		// documentIds = append(documentIds, oldDoc.ID)
		// 创建新文档记录
		newDoc := models.Document{
			Id:        strconv.FormatInt(oldDoc.ID, 10),
			CreatedAt: oldDoc.CreatedAt,
			UpdatedAt: oldDoc.UpdatedAt,
			DeletedAt: models.DeletedAt{},
			UserId:    getUserID(oldDoc.UserId),
			Path:      oldDoc.Path,
			DocType:   models.DocType(oldDoc.DocType),
			Name:      oldDoc.Name,
			Size:      oldDoc.Size,
			DeleteBy: func() string {
				if oldDoc.DeleteBy == 0 {
					return ""
				}
				return getUserID(oldDoc.DeleteBy)
			}(),
			VersionId: oldDoc.VersionId,
			TeamId: func() string {
				if oldDoc.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldDoc.TeamId, 10)
			}(),
			ProjectId: func() string {
				if oldDoc.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldDoc.ProjectId, 10)
			}(),
		}

		// 设置DeletedAt
		if oldDoc.DeletedAt != nil {
			newDoc.DeletedAt.Time = *oldDoc.DeletedAt
			newDoc.DeletedAt.Valid = true
		}

		// 检查并更新文档记录
		if err := checkAndUpdate(targetDB, "document", "id = ?", newDoc.Id, newDoc); err != nil {
			log.Printf("Error migrating document %d: %v", oldDoc.ID, err)
			continue
		}

		// 迁移文档数据
		err := migrateDocumentStorage(oldDoc.ID, config.Source.GenerateApiUrl)
		if err != nil {
			log.Println("migrateDocumentStorage failed", err)
			continue
		}

		// 如果有锁定信息，检查并更新DocumentLock记录
		if !oldDoc.LockedAt.IsZero() || oldDoc.LockedReason != "" || oldDoc.LockedWords != "" {
			docLock := models.DocumentLock{
				DocumentId:   newDoc.Id,
				LockedReason: oldDoc.LockedReason,
				LockedType:   models.LockedTypeText,
				LockedTarget: "",
				LockedWords:  oldDoc.LockedWords,
			}
			if err := checkAndUpdate(targetDB, "document_lock", "document_id = ?", docLock.DocumentId, docLock); err != nil {
				log.Printf("Error creating/updating document lock for document %s: %v", newDoc.Id, err)
			}
		}
	}
	// 迁移文档权限申请表
	var oldDocPermRequests []struct {
		ID               int64      `gorm:"column:id"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		DeletedAt        *time.Time `gorm:"column:deleted_at"`
		UserId           int64      `gorm:"column:user_id"`
		DocumentId       int64      `gorm:"column:document_id"`
		PermType         uint8      `gorm:"column:perm_type"`
		Status           uint8      `gorm:"column:status"`
		FirstDisplayedAt time.Time  `gorm:"column:first_displayed_at"`
		ProcessedAt      time.Time  `gorm:"column:processed_at"`
		ProcessedBy      int64      `gorm:"column:processed_by"`
		ApplicantNotes   string     `gorm:"column:applicant_notes"`
		ProcessorNotes   string     `gorm:"column:processor_notes"`
	}

	if err := sourceDB.Table("document_permission_requests").Find(&oldDocPermRequests).Error; err != nil {
		log.Fatalf("Error querying document permission requests: %v", err)
	}

	for _, oldRequest := range oldDocPermRequests {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldRequest.FirstDisplayedAt)
		customProcessedAt := utilTime.Time(oldRequest.ProcessedAt)

		newRequest := models.DocumentPermissionRequests{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRequest.CreatedAt,
				UpdatedAt: oldRequest.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:           getUserID(oldRequest.UserId),
			DocumentId:       strconv.FormatInt(oldRequest.DocumentId, 10),
			PermType:         models.PermType(oldRequest.PermType),
			Status:           models.StatusType(oldRequest.Status),
			FirstDisplayedAt: customFirstDisplayedAt,
			ProcessedAt:      customProcessedAt,
			ProcessedBy: func() string {
				if oldRequest.ProcessedBy == 0 {
					return ""
				}
				return getUserID(oldRequest.ProcessedBy)
			}(),
			ApplicantNotes: oldRequest.ApplicantNotes,
			ProcessorNotes: oldRequest.ProcessorNotes,
		}

		if oldRequest.DeletedAt != nil {
			newRequest.DeletedAt.Time = *oldRequest.DeletedAt
			newRequest.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "document_permission_requests", "document_id = ? AND user_id = ? AND perm_type = ?",
			[]interface{}{newRequest.DocumentId, newRequest.UserId, newRequest.PermType}, newRequest); err != nil {
			log.Printf("Error migrating document permission request %d: %v", oldRequest.ID, err)
		}
	}

	// 迁移文档版本表
	var oldVersions []struct {
		DeletedAt  *time.Time `gorm:"column:deleted_at"`
		CreatedAt  time.Time  `gorm:"column:created_at"`
		UpdatedAt  time.Time  `gorm:"column:updated_at"`
		ID         int64      `gorm:"column:id"`
		DocumentId int64      `gorm:"column:document_id"`
		VersionId  string     `gorm:"column:version_id"`
		LastCmdId  int64      `gorm:"column:last_cmd_id"`
		// 其他BaseModel字段
	}

	if err := sourceDB.Table("document_version").Find(&oldVersions).Error; err != nil {
		log.Fatalf("Error querying document versions: %v", err)
	}
	for _, oldVer := range oldVersions {
		newVer := models.DocumentVersion{
			BaseModelStruct: models.BaseModelStruct{
				DeletedAt: models.DeletedAt{},
				CreatedAt: oldVer.CreatedAt,
				UpdatedAt: oldVer.UpdatedAt,
			},
			DocumentId:   strconv.FormatInt(oldVer.DocumentId, 10), // 转为string
			VersionId:    oldVer.VersionId,
			LastCmdVerId: uint(oldVer.LastCmdId), // 注意这里字段名和类型都改变
		}

		if oldVer.DeletedAt != nil {
			newVer.DeletedAt.Time = *oldVer.DeletedAt
			newVer.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "document_version", "document_id = ? AND version_id = ?", []interface{}{newVer.DocumentId, newVer.VersionId}, newVer); err != nil {
			log.Printf("Error migrating version %d: %v", oldVer.ID, err)
		}
	}

	// 迁移文档权限表
	var oldPermissions []struct {
		ID             int64      `gorm:"column:id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		DeletedAt      *time.Time `gorm:"column:deleted_at"`
		ResourceType   uint8      `gorm:"column:resource_type"`
		ResourceId     int64      `gorm:"column:resource_id"`
		GranteeType    uint8      `gorm:"column:grantee_type"`
		GranteeId      int64      `gorm:"column:grantee_id"`
		PermType       uint8      `gorm:"column:perm_type"`
		PermSourceType uint8      `gorm:"column:perm_source_type"`
	}

	if err := sourceDB.Table("document_permission").Find(&oldPermissions).Error; err != nil {
		log.Fatalf("Error querying document permissions: %v", err)
	}

	for _, oldPerm := range oldPermissions {
		newPerm := models.DocumentPermission{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldPerm.CreatedAt,
				UpdatedAt: oldPerm.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			ResourceType:   models.ResourceType(oldPerm.ResourceType),
			ResourceId:     strconv.FormatInt(oldPerm.ResourceId, 10),
			GranteeType:    models.GranteeType(oldPerm.GranteeType),
			GranteeId:      strconv.FormatInt(oldPerm.GranteeId, 10),
			PermType:       models.PermType(oldPerm.PermType),
			PermSourceType: models.PermSourceType(oldPerm.PermSourceType),
		}

		if oldPerm.DeletedAt != nil {
			newPerm.DeletedAt.Time = *oldPerm.DeletedAt
			newPerm.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "document_permission", "resource_id = ? AND grantee_id = ?", []interface{}{newPerm.ResourceId, newPerm.GranteeId}, newPerm); err != nil {
			log.Printf("Error migrating permission %d: %v", oldPerm.ID, err)
		}
	}

	// 迁移文档访问记录表
	var oldAccessRecords []struct {
		ID             int64      `gorm:"column:id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		DeletedAt      *time.Time `gorm:"column:deleted_at"`
		UserId         int64      `gorm:"column:user_id"`
		DocumentId     int64      `gorm:"column:document_id"`
		LastAccessTime time.Time  `gorm:"column:last_access_time"`
	}

	if err := sourceDB.Table("document_access_record").Find(&oldAccessRecords).Error; err != nil {
		log.Fatalf("Error querying document access records: %v", err)
	}

	for _, oldRecord := range oldAccessRecords {
		newRecord := models.DocumentAccessRecord{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRecord.CreatedAt,
				UpdatedAt: oldRecord.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:         getUserID(oldRecord.UserId),
			DocumentId:     strconv.FormatInt(oldRecord.DocumentId, 10),
			LastAccessTime: oldRecord.LastAccessTime,
		}

		if oldRecord.DeletedAt != nil {
			newRecord.DeletedAt.Time = *oldRecord.DeletedAt
			newRecord.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "document_access_record", "user_id = ? AND document_id = ?", []interface{}{newRecord.UserId, newRecord.DocumentId}, newRecord); err != nil {
			log.Printf("Error migrating access record %d: %v", oldRecord.ID, err)
		}
	}

	// 迁移文档收藏表
	var oldFavorites []struct {
		ID         int64      `gorm:"column:id"`
		CreatedAt  time.Time  `gorm:"column:created_at"`
		UpdatedAt  time.Time  `gorm:"column:updated_at"`
		DeletedAt  *time.Time `gorm:"column:deleted_at"`
		UserId     int64      `gorm:"column:user_id"`
		DocumentId int64      `gorm:"column:document_id"`
		IsFavorite bool       `gorm:"column:is_favorite"`
	}

	if err := sourceDB.Table("document_favorites").Find(&oldFavorites).Error; err != nil {
		log.Fatalf("Error querying document favorites: %v", err)
	}
	for _, oldFav := range oldFavorites {
		newFav := models.DocumentFavorites{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldFav.CreatedAt,
				UpdatedAt: oldFav.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:     getUserID(oldFav.UserId),
			DocumentId: strconv.FormatInt(oldFav.DocumentId, 10),
			IsFavorite: oldFav.IsFavorite,
		}

		if oldFav.DeletedAt != nil {
			newFav.DeletedAt.Time = *oldFav.DeletedAt
			newFav.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "document_favorites", "user_id = ? AND document_id = ?", []interface{}{newFav.UserId, newFav.DocumentId}, newFav); err != nil {
			log.Printf("Error migrating favorite %d: %v", oldFav.ID, err)
		}
	}

	// 迁移团队表
	var oldTeams []struct {
		ID              int64      `gorm:"column:id"`
		CreatedAt       time.Time  `gorm:"column:created_at"`
		UpdatedAt       time.Time  `gorm:"column:updated_at"`
		DeletedAt       *time.Time `gorm:"column:deleted_at"`
		Name            string     `gorm:"column:name"`
		Description     string     `gorm:"column:description"`
		Avatar          string     `gorm:"column:avatar"`
		Uid             string     `gorm:"column:uid"`
		InvitedPermType uint8      `gorm:"column:invited_perm_type"`
		InvitedSwitch   bool       `gorm:"column:invited_switch"`
	}

	if err := sourceDB.Table("team").Find(&oldTeams).Error; err != nil {
		log.Fatalf("Error querying teams: %v", err)
	}

	for _, oldTeam := range oldTeams {
		// 转换时间类型
		customCreatedAt := utilTime.Time(oldTeam.CreatedAt)
		customUpdatedAt := utilTime.Time(oldTeam.UpdatedAt)

		newTeam := models.Team{
			Id:              strconv.FormatInt(oldTeam.ID, 10),
			CreatedAt:       customCreatedAt,
			UpdatedAt:       customUpdatedAt,
			DeletedAt:       models.DeletedAt{},
			Name:            oldTeam.Name,
			Description:     oldTeam.Description,
			Avatar:          oldTeam.Avatar,
			InvitedPermType: models.TeamPermType(oldTeam.InvitedPermType),
			OpenInvite:      oldTeam.InvitedSwitch,
		}

		if oldTeam.DeletedAt != nil {
			newTeam.DeletedAt.Time = *oldTeam.DeletedAt
			newTeam.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "team", "id = ?", newTeam.Id, newTeam); err != nil {
			log.Printf("Error migrating team %d: %v", oldTeam.ID, err)
		}
	}

	// 迁移团队成员表
	var oldTeamMembers []struct {
		ID        int64      `gorm:"column:id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		TeamId    int64      `gorm:"column:team_id"`
		UserId    int64      `gorm:"column:user_id"`
		PermType  uint8      `gorm:"column:perm_type"`
		Nickname  string     `gorm:"column:nickname"`
	}

	if err := sourceDB.Table("team_member").Find(&oldTeamMembers).Error; err != nil {
		log.Fatalf("Error querying team members: %v", err)
	}

	for _, oldMember := range oldTeamMembers {
		newMember := models.TeamMember{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMember.CreatedAt,
				UpdatedAt: oldMember.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			TeamId: func() string {
				if oldMember.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMember.TeamId, 10)
			}(),
			UserId:   getUserID(oldMember.UserId),
			PermType: models.TeamPermType(oldMember.PermType),
			Nickname: oldMember.Nickname,
		}

		if oldMember.DeletedAt != nil {
			newMember.DeletedAt.Time = *oldMember.DeletedAt
			newMember.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "team_member", "team_id = ? AND user_id = ?", []interface{}{newMember.TeamId, newMember.UserId}, newMember); err != nil {
			log.Printf("Error migrating team member %d: %v", oldMember.ID, err)
		}
	}

	// 迁移团队加入申请表
	var oldTeamJoinRequests []struct {
		ID               int64      `gorm:"column:id"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		DeletedAt        *time.Time `gorm:"column:deleted_at"`
		UserId           int64      `gorm:"column:user_id"`
		TeamId           int64      `gorm:"column:team_id"`
		PermType         uint8      `gorm:"column:perm_type"`
		Status           uint8      `gorm:"column:status"`
		FirstDisplayedAt time.Time  `gorm:"column:first_displayed_at"`
		ProcessedAt      time.Time  `gorm:"column:processed_at"`
		ProcessedBy      int64      `gorm:"column:processed_by"`
		ApplicantNotes   string     `gorm:"column:applicant_notes"`
		ProcessorNotes   string     `gorm:"column:processor_notes"`
	}

	if err := sourceDB.Table("team_join_request").Find(&oldTeamJoinRequests).Error; err != nil {
		log.Fatalf("Error querying team join requests: %v", err)
	}

	for _, oldRequest := range oldTeamJoinRequests {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldRequest.FirstDisplayedAt)
		customProcessedAt := utilTime.Time(oldRequest.ProcessedAt)

		newRequest := models.TeamJoinRequest{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRequest.CreatedAt,
				UpdatedAt: oldRequest.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:           getUserID(oldRequest.UserId),
			TeamId:           strconv.FormatInt(oldRequest.TeamId, 10),
			PermType:         models.TeamPermType(oldRequest.PermType),
			Status:           models.TeamJoinRequestStatus(oldRequest.Status),
			FirstDisplayedAt: customFirstDisplayedAt,
			ProcessedAt:      customProcessedAt,
			ProcessedBy:      getUserID(oldRequest.ProcessedBy),
			ApplicantNotes:   oldRequest.ApplicantNotes,
			ProcessorNotes:   oldRequest.ProcessorNotes,
		}

		// 处理空ID
		if oldRequest.ProcessedBy == 0 {
			newRequest.ProcessedBy = ""
		}

		// 设置DeletedAt
		if oldRequest.DeletedAt != nil {
			newRequest.DeletedAt.Time = *oldRequest.DeletedAt
			newRequest.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "team_join_request", "user_id = ? AND team_id = ?", []interface{}{newRequest.UserId, newRequest.TeamId}, newRequest); err != nil {
			log.Printf("Error migrating team join request %d: %v", oldRequest.ID, err)
		}
	}

	// 迁移团队加入申请消息表
	var oldTeamMessageShows []struct {
		ID                int64      `gorm:"column:id"`
		CreatedAt         time.Time  `gorm:"column:created_at"`
		UpdatedAt         time.Time  `gorm:"column:updated_at"`
		DeletedAt         *time.Time `gorm:"column:deleted_at"`
		TeamJoinRequestId int64      `gorm:"column:team_join_request_id"`
		UserId            int64      `gorm:"column:user_id"`
		TeamId            int64      `gorm:"column:team_id"`
		FirstDisplayedAt  time.Time  `gorm:"column:first_displayed_at"`
	}

	if err := sourceDB.Table("team_join_request_message_show").Find(&oldTeamMessageShows).Error; err != nil {
		log.Fatalf("Error querying team join request messages: %v", err)
	}

	for _, oldMessage := range oldTeamMessageShows {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldMessage.FirstDisplayedAt)

		newMessage := models.TeamJoinRequestMessageShow{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMessage.CreatedAt,
				UpdatedAt: oldMessage.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			// TeamJoinRequestId 保持 int64 类型
			TeamJoinRequestId: oldMessage.TeamJoinRequestId,
			// 转换 UserId 和 TeamId 为 string 类型
			UserId: strconv.FormatInt(oldMessage.UserId, 10),
			TeamId: func() string {
				if oldMessage.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMessage.TeamId, 10)
			}(),
			FirstDisplayedAt: customFirstDisplayedAt,
		}

		if oldMessage.DeletedAt != nil {
			newMessage.DeletedAt.Time = *oldMessage.DeletedAt
			newMessage.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "team_join_request_message_show", "team_join_request_id = ? AND team_id = ?", []interface{}{newMessage.TeamJoinRequestId, newMessage.TeamId}, newMessage); err != nil {
			log.Printf("Error migrating team join request message %d: %v", oldMessage.ID, err)
		}
	}

	// 迁移项目表
	var oldProjects []struct {
		ID            int64      `gorm:"column:id"`
		CreatedAt     time.Time  `gorm:"column:created_at"`
		UpdatedAt     time.Time  `gorm:"column:updated_at"`
		DeletedAt     *time.Time `gorm:"column:deleted_at"`
		TeamId        int64      `gorm:"column:team_id"`
		Name          string     `gorm:"column:name"`
		Description   string     `gorm:"column:description"`
		PublicSwitch  bool       `gorm:"column:public_switch"`
		PermType      uint8      `gorm:"column:perm_type"`
		InvitedSwitch bool       `gorm:"column:invited_switch"`
		NeedApproval  bool       `gorm:"column:need_approval"`
	}

	if err := sourceDB.Table("project").Find(&oldProjects).Error; err != nil {
		log.Fatalf("Error querying projects: %v", err)
	}

	for _, oldProject := range oldProjects {
		// 转换时间类型
		customCreatedAt := utilTime.Time(oldProject.CreatedAt)
		customUpdatedAt := utilTime.Time(oldProject.UpdatedAt)

		newProject := models.Project{
			Id:        strconv.FormatInt(oldProject.ID, 10),
			CreatedAt: customCreatedAt,
			UpdatedAt: customUpdatedAt,
			DeletedAt: models.DeletedAt{},
			TeamId: func() string {
				if oldProject.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldProject.TeamId, 10)
			}(),
			Name:         oldProject.Name,
			Description:  oldProject.Description,
			IsPublic:     oldProject.PublicSwitch,
			PermType:     models.ProjectPermType(oldProject.PermType),
			OpenInvite:   oldProject.InvitedSwitch,
			NeedApproval: oldProject.NeedApproval,
		}

		if oldProject.DeletedAt != nil {
			newProject.DeletedAt.Time = *oldProject.DeletedAt
			newProject.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "project", "id = ?", newProject.Id, newProject); err != nil {
			log.Printf("Error migrating project %d: %v", oldProject.ID, err)
		}
	}

	// 迁移项目收藏表
	var oldProjectFavorites []struct {
		ID        int64      `gorm:"column:id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		UserId    int64      `gorm:"column:user_id"`
		ProjectId int64      `gorm:"column:project_id"`
		IsFavor   bool       `gorm:"column:is_favor"`
	}

	if err := sourceDB.Table("project_favorite").Find(&oldProjectFavorites).Error; err != nil {
		log.Fatalf("Error querying project favorites: %v", err)
	}

	for _, oldFav := range oldProjectFavorites {
		newFav := models.ProjectFavorite{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldFav.CreatedAt,
				UpdatedAt: oldFav.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:    getUserID(oldFav.UserId),
			ProjectId: strconv.FormatInt(oldFav.ProjectId, 10),
			IsFavor:   oldFav.IsFavor,
		}

		if oldFav.DeletedAt != nil {
			newFav.DeletedAt.Time = *oldFav.DeletedAt
			newFav.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "project_favorite", "user_id = ? AND project_id = ?", []interface{}{newFav.UserId, newFav.ProjectId}, newFav); err != nil {
			log.Printf("Error migrating project favorite %d: %v", oldFav.ID, err)
		}
	}

	// 迁移项目申请表
	var oldProjectJoinRequests []struct {
		ID               int64      `gorm:"column:id"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		DeletedAt        *time.Time `gorm:"column:deleted_at"`
		UserId           int64      `gorm:"column:user_id"`
		ProjectId        int64      `gorm:"column:project_id"`
		PermType         uint8      `gorm:"column:perm_type"`
		Status           uint8      `gorm:"column:status"`
		FirstDisplayedAt time.Time  `gorm:"column:first_displayed_at"`
		ProcessedAt      time.Time  `gorm:"column:processed_at"`
		ProcessedBy      int64      `gorm:"column:processed_by"`
		ApplicantNotes   string     `gorm:"column:applicant_notes"`
		ProcessorNotes   string     `gorm:"column:processor_notes"`
	}

	if err := sourceDB.Table("project_join_request").Find(&oldProjectJoinRequests).Error; err != nil {
		log.Fatalf("Error querying project join requests: %v", err)
	}

	for _, oldRequest := range oldProjectJoinRequests {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldRequest.FirstDisplayedAt)
		customProcessedAt := utilTime.Time(oldRequest.ProcessedAt)

		newRequest := models.ProjectJoinRequest{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRequest.CreatedAt,
				UpdatedAt: oldRequest.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId: strconv.FormatInt(oldRequest.UserId, 10),
			ProjectId: func() string {
				if oldRequest.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldRequest.ProjectId, 10)
			}(),
			PermType:         models.ProjectPermType(oldRequest.PermType),
			Status:           models.ProjectJoinRequestStatus(oldRequest.Status),
			FirstDisplayedAt: customFirstDisplayedAt,
			ProcessedAt:      customProcessedAt,
			ProcessedBy:      getUserID(oldRequest.ProcessedBy),
			ApplicantNotes:   oldRequest.ApplicantNotes,
			ProcessorNotes:   oldRequest.ProcessorNotes,
		}

		// 处理空ID
		if oldRequest.ProcessedBy == 0 {
			newRequest.ProcessedBy = ""
		}

		// 设置DeletedAt
		if oldRequest.DeletedAt != nil {
			newRequest.DeletedAt.Time = *oldRequest.DeletedAt
			newRequest.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "project_join_request", "user_id = ? AND project_id = ?", []interface{}{newRequest.UserId, newRequest.ProjectId}, newRequest); err != nil {
			log.Printf("Error migrating project join request %d: %v", oldRequest.ID, err)
		}
	}

	// 迁移项目申请消息表
	var oldMessageShows []struct {
		ID                   int64      `gorm:"column:id"`
		CreatedAt            time.Time  `gorm:"column:created_at"`
		UpdatedAt            time.Time  `gorm:"column:updated_at"`
		DeletedAt            *time.Time `gorm:"column:deleted_at"`
		ProjectJoinRequestId int64      `gorm:"column:project_join_request_id"`
		UserId               int64      `gorm:"column:user_id"`
		ProjectId            int64      `gorm:"column:project_id"`
		FirstDisplayedAt     time.Time  `gorm:"column:first_displayed_at"`
	}

	if err := sourceDB.Table("project_join_request_message_show").Find(&oldMessageShows).Error; err != nil {
		log.Fatalf("Error querying project join request messages: %v", err)
	}

	for _, oldMessage := range oldMessageShows {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldMessage.FirstDisplayedAt)

		newMessage := models.ProjectJoinRequestMessageShow{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMessage.CreatedAt,
				UpdatedAt: oldMessage.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			// 保持 ProjectJoinRequestId 为 int64 类型
			ProjectJoinRequestId: oldMessage.ProjectJoinRequestId,
			// 转换 UserId 和 ProjectId 为 string 类型
			UserId: strconv.FormatInt(oldMessage.UserId, 10),
			ProjectId: func() string {
				if oldMessage.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMessage.ProjectId, 10)
			}(),
			FirstDisplayedAt: customFirstDisplayedAt,
		}

		if oldMessage.DeletedAt != nil {
			newMessage.DeletedAt.Time = *oldMessage.DeletedAt
			newMessage.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "project_join_request_message_show", "project_join_request_id = ? AND project_id = ?", []interface{}{newMessage.ProjectJoinRequestId, newMessage.ProjectId}, newMessage); err != nil {
			log.Printf("Error migrating project join request message %d: %v", oldMessage.ID, err)
		}
	}

	// 迁移项目成员表
	var oldProjectMembers []struct {
		ID             int64      `gorm:"column:id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		DeletedAt      *time.Time `gorm:"column:deleted_at"`
		ProjectId      int64      `gorm:"column:project_id"`
		UserId         int64      `gorm:"column:user_id"`
		PermType       uint8      `gorm:"column:perm_type"`
		PermSourceType uint8      `gorm:"column:perm_source_type"`
	}

	if err := sourceDB.Table("project_member").Find(&oldProjectMembers).Error; err != nil {
		log.Fatalf("Error querying project members: %v", err)
	}

	for _, oldMember := range oldProjectMembers {
		newMember := models.ProjectMember{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMember.CreatedAt,
				UpdatedAt: oldMember.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			ProjectId: func() string {
				if oldMember.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMember.ProjectId, 10)
			}(),
			UserId:         getUserID(oldMember.UserId),
			PermType:       models.ProjectPermType(oldMember.PermType),
			PermSourceType: models.ProjectPermSourceType(oldMember.PermSourceType),
		}

		if oldMember.DeletedAt != nil {
			newMember.DeletedAt.Time = *oldMember.DeletedAt
			newMember.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "project_member", "project_id = ? AND user_id = ?", []interface{}{newMember.ProjectId, newMember.UserId}, newMember); err != nil {
			log.Printf("Error migrating project member %d: %v", oldMember.ID, err)
		}
	}

	// 迁移反馈表
	var oldFeedbacks []struct {
		ID            int64      `gorm:"column:id"`
		CreatedAt     time.Time  `gorm:"column:created_at"`
		UpdatedAt     time.Time  `gorm:"column:updated_at"`
		DeletedAt     *time.Time `gorm:"column:deleted_at"`
		UserId        int64      `gorm:"column:user_id"`
		Type          uint8      `gorm:"column:type"`
		Content       string     `gorm:"column:content"`
		ImagePathList string     `gorm:"column:image_path_list"`
		PageUrl       string     `gorm:"column:page_url"`
	}

	if err := sourceDB.Table("feedback").Find(&oldFeedbacks).Error; err != nil {
		log.Fatalf("Error querying feedbacks: %v", err)
	}

	for _, oldFeedback := range oldFeedbacks {
		newFeedback := models.Feedback{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldFeedback.CreatedAt,
				UpdatedAt: oldFeedback.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:        getUserID(oldFeedback.UserId),
			Type:          models.FeedbackType(oldFeedback.Type),
			Content:       oldFeedback.Content,
			ImagePathList: oldFeedback.ImagePathList,
			PageUrl:       oldFeedback.PageUrl,
		}

		if oldFeedback.DeletedAt != nil {
			newFeedback.DeletedAt.Time = *oldFeedback.DeletedAt
			newFeedback.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "feedback", "user_id = ?", []interface{}{newFeedback.UserId}, newFeedback); err != nil {
			log.Printf("Error migrating feedback %d: %v", oldFeedback.ID, err)
		}
	}

	// 迁移用户键值存储表
	var oldUserKVStorages []struct {
		ID        int64      `gorm:"column:id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		UserId    int64      `gorm:"column:user_id"`
		Key       string     `gorm:"column:key"`
		Value     string     `gorm:"column:value"`
	}

	if err := sourceDB.Table("user_kv_storage").Find(&oldUserKVStorages).Error; err != nil {
		log.Fatalf("Error querying user kv storages: %v", err)
	}

	for _, oldKV := range oldUserKVStorages {
		newKV := models.UserKVStorage{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldKV.CreatedAt,
				UpdatedAt: oldKV.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId: getUserID(oldKV.UserId),
			Key:    oldKV.Key,
			Value:  oldKV.Value,
		}

		if oldKV.DeletedAt != nil {
			newKV.DeletedAt.Time = *oldKV.DeletedAt
			newKV.DeletedAt.Valid = true
		}

		if err := checkAndUpdate(targetDB, "user_kv_storage", "user_id = ? AND `key` = ?", []interface{}{newKV.UserId, newKV.Key}, newKV); err != nil {
			log.Printf("Error migrating user kv storage %d: %v", oldKV.ID, err)
		}
	}

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
					newComment["user"] = getUserID(oldId)
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
	// 直接插入评论，不检测是否已有插入的数据
	// if len(newComments) > 0 {
	// 	log.Printf("Inserting %d comments", len(newComments))
	// 	if _, err := targetMongo.DB.Collection("comment").InsertMany(context.Background(), newComments); err != nil {
	// 		log.Printf("Error inserting comments: %v", err)
	// 	}
	// }
	// migrateDocumentStorage(documentIds, config.Source.GenerateApiUrl)

	log.Println("Migration completed!")
}

// 辅助函数：检查记录是否存在并更新
func checkAndUpdate(db *gorm.DB, table string, whereClause string, whereArgs interface{}, newRecord interface{}) error {
	var count int64
	query := db.Table(table)

	// 处理多个参数的情况
	if args, ok := whereArgs.([]interface{}); ok {
		query = query.Where(whereClause, args...)
	} else {
		query = query.Where(whereClause, whereArgs)
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		// 记录存在，执行更新
		if err := query.Updates(newRecord).Error; err != nil {
			return err
		}
		log.Printf("Updated existing record in %s with %s = %v", table, whereClause, whereArgs)
	} else {
		// 记录不存在，执行创建
		recordPtr := reflect.New(reflect.TypeOf(newRecord)).Interface()
		reflect.ValueOf(recordPtr).Elem().Set(reflect.ValueOf(newRecord))
		if err := db.Table(table).Create(recordPtr).Error; err != nil {
			return err
		}
		log.Printf("Created new record in %s with %s = %v", table, whereClause, whereArgs)
	}
	return nil
}
