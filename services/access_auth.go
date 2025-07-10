package services

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"kcaitech.com/kcserver/models"
)

type AccessAuthService struct {
	*DefaultService
}

func NewAccessAuthService() *AccessAuthService {
	that := &AccessAuthService{
		DefaultService: NewDefaultService(&models.AccessAuth{}),
	}
	that.That = that
	return that
}

func (s *AccessAuthService) GetAccessAuth(accessKey string) (*models.AccessAuth, error) {
	var accessAuth models.AccessAuth
	if err := s.DBModule.DB.Where("key = ?", accessKey).First(&accessAuth).Error; err != nil {
		return nil, err
	}
	return &accessAuth, nil
}

func (s *AccessAuthService) GetAccessAuthByUserId(userId string) ([]*models.AccessAuth, error) {
	var accessAuths []*models.AccessAuth
	if err := s.DBModule.DB.Where("user_id = ?", userId).Find(&accessAuths).Error; err != nil {
		return nil, err
	}
	return accessAuths, nil
}

type CombinedAccessAuth struct {
	*models.AccessAuth
	*models.AccessAuthResource
}

func (s *AccessAuthService) GetCombinedAccessAuth(userId string) ([]*CombinedAccessAuth, error) {
	var combinedAccessAuths []*CombinedAccessAuth

	// 使用JOIN查询获取AccessAuth和AccessAuthResource的组合数据
	if err := s.DBModule.DB.
		Table("access_auth").
		Select("access_auth.*, access_auth_resource.*").
		Joins("LEFT JOIN access_auth_resource ON access_auth.key = access_auth_resource.key").
		Where("access_auth.user_id = ?", userId).
		Scan(&combinedAccessAuths).Error; err != nil {
		return nil, err
	}

	return combinedAccessAuths, nil
}

func (s *AccessAuthService) CreateAccessAuth(accessAuth *models.AccessAuth) error {
	return s.DBModule.DB.Create(accessAuth).Error
}

func (s *AccessAuthService) SaveAccessAuth(accessAuth *models.AccessAuth) error {
	return s.DBModule.DB.Save(accessAuth).Error
}

func (s *AccessAuthService) DeleteAccessAuth(accessKey string) error {
	return s.DBModule.DB.Where("key = ?", accessKey).Delete(&models.AccessAuth{}).Error
}

func (s *AccessAuthService) GetAccessAuthResource(key string) ([]models.AccessAuthResource, error) {
	var accessAuthRanges []models.AccessAuthResource
	if err := s.DBModule.DB.Where("key = ?", key).Find(&accessAuthRanges).Error; err != nil {
		return nil, err
	}
	return accessAuthRanges, nil
}

func (s *AccessAuthService) CreateAccessAuthResource(accessAuthRange *models.AccessAuthResource) error {
	return s.DBModule.DB.Create(accessAuthRange).Error
}

func (s *AccessAuthService) SaveAccessAuthResource(accessAuthRange *models.AccessAuthResource) error {
	return s.DBModule.DB.Save(accessAuthRange).Error
}

func (s *AccessAuthService) DeleteAccessAuthResourceNotExists(key string, resourceType uint8, resourceIds []string) error {
	return s.DBModule.DB.Where("key = ? AND type = ? AND resource_id NOT IN (?)", key, resourceType, resourceIds).Delete(&models.AccessAuthResource{}).Error
}

func (s *AccessAuthService) DeleteAccessAuthResource(key string, resourceType uint8, resourceId string) error {
	return s.DBModule.DB.Where("key = ? AND type = ? AND resource_id = ?", key, resourceType, resourceId).Delete(&models.AccessAuthResource{}).Error
}

type GrantPost struct {
	Priority []string `json:"priority"` // read, comment, write, delete
	Document []string `json:"document"` // 文档id
	Project  []string `json:"project"`  // 项目id
	Team     []string `json:"team"`     // 团队id
	User     bool     `json:"user"`     // 当前用户
}

func HashPassword(password string) string {
	cost := bcrypt.DefaultCost

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Fatalf("生成密码哈希时出错: %v", err)
	}

	return string(hashedPassword)
}

func (s *AccessAuthService) UpdateAccessAuth(userId string, accessKey string, accessSecret string, grantPost GrantPost) error {
	priorityMask := 0
	for _, priority := range grantPost.Priority {
		switch priority {
		case "read":
			priorityMask |= models.AccessAuthPriorityMaskRead
		case "comment":
			priorityMask |= models.AccessAuthPriorityMaskComment
		case "write":
			priorityMask |= models.AccessAuthPriorityMaskWrite
		case "delete":
			priorityMask |= models.AccessAuthPriorityMaskDelete
		}
	}

	resourceMask := 0
	if grantPost.User {
		resourceMask |= models.AccessAuthResourceMaskUser
	}
	if len(grantPost.Document) > 0 {
		resourceMask |= models.AccessAuthResourceMaskDocument
	}
	if len(grantPost.Project) > 0 {
		resourceMask |= models.AccessAuthResourceMaskProject
	}
	if len(grantPost.Team) > 0 {
		resourceMask |= models.AccessAuthResourceMaskTeam
	}

	accessAuth := &models.AccessAuth{
		UserId:       userId,
		Key:          accessKey,
		PriorityMask: uint32(priorityMask),
		ResourceMask: uint32(resourceMask),
	}

	if accessSecret != "" {
		accessAuth.Secret = HashPassword(accessSecret)
	}

	if err := s.SaveAccessAuth(accessAuth); err != nil {
		return err
	}

	for _, document := range grantPost.Document {
		accessAuthRange := &models.AccessAuthResource{
			Key:        accessKey,
			ResourceId: document,
			Type:       uint8(models.AccessAuthResourceTypeDocument),
		}
		if err := s.CreateAccessAuthResource(accessAuthRange); err != nil {
			return err
		}
	}

	for _, project := range grantPost.Project {
		accessAuthRange := &models.AccessAuthResource{
			Key:        accessKey,
			ResourceId: project,
			Type:       uint8(models.AccessAuthResourceTypeProject),
		}
		if err := s.CreateAccessAuthResource(accessAuthRange); err != nil {
			return err
		}
	}

	for _, team := range grantPost.Team {
		accessAuthRange := &models.AccessAuthResource{
			Key:        accessKey,
			ResourceId: team,
			Type:       uint8(models.AccessAuthResourceTypeTeam),
		}
		if err := s.CreateAccessAuthResource(accessAuthRange); err != nil {
			return err
		}
	}

	return nil
}
