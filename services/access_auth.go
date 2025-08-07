/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package services

import (
	"encoding/json"
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
	if err := s.DBModule.DB.Where("access_key = ?", accessKey).First(&accessAuth).Error; err != nil {
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
	// AccessAuth 字段
	UserId       string `json:"user_id"`
	AccessKey    string `json:"key"`
	PriorityMask uint32 `json:"priority_mask"`
	ResourceMask uint32 `json:"resource_mask"`
	// AccessAuthResource 字段
	ResourceId string `json:"resource_id"`
	Type       uint8  `json:"type"`
}

func (c *CombinedAccessAuth) MarshalJSON() ([]byte, error) {
	type Alias CombinedAccessAuth
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

func (s *AccessAuthService) GetCombinedAccessAuth(userId string) ([]*CombinedAccessAuth, error) {
	var combinedAccessAuths []*CombinedAccessAuth = make([]*CombinedAccessAuth, 0)

	// 使用JOIN查询获取AccessAuth和AccessAuthResource的组合数据
	// 添加软删除条件，过滤已删除的记录
	if err := s.DBModule.DB.
		Table("access_auth").
		Select("access_auth.user_id, access_auth.access_key, access_auth.priority_mask, access_auth.resource_mask, access_auth_resource.type, access_auth_resource.resource_id").
		Joins("LEFT JOIN access_auth_resource ON access_auth.access_key = access_auth_resource.access_key AND access_auth_resource.deleted_at IS NULL").
		Where("access_auth.user_id = ? AND access_auth.deleted_at IS NULL", userId).
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
	return s.DBModule.DB.Where("access_key = ?", accessKey).Delete(&models.AccessAuth{}).Error
}

func (s *AccessAuthService) GetAccessAuthResource(key string) ([]models.AccessAuthResource, error) {
	var accessAuthResources []models.AccessAuthResource
	if err := s.DBModule.DB.Where("access_key = ?", key).Find(&accessAuthResources).Error; err != nil {
		return nil, err
	}
	return accessAuthResources, nil
}

func (s *AccessAuthService) CreateAccessAuthResource(accessAuthResource *models.AccessAuthResource) error {
	return s.DBModule.DB.Create(accessAuthResource).Error
}

func (s *AccessAuthService) SaveAccessAuthResource(accessAuthResource *models.AccessAuthResource) error {
	return s.DBModule.DB.Save(accessAuthResource).Error
}

func (s *AccessAuthService) DeleteAccessAuthResourceNotExists(key string, resourceType uint8, resourceIds []string) error {
	return s.DBModule.DB.Where("access_key = ? AND type = ? AND resource_id NOT IN (?)", key, resourceType, resourceIds).Delete(&models.AccessAuthResource{}).Error
}

func (s *AccessAuthService) DeleteAccessAuthResource(key string, resourceType uint8, resourceId string) error {
	return s.DBModule.DB.Where("access_key = ? AND type = ? AND resource_id = ?", key, resourceType, resourceId).Delete(&models.AccessAuthResource{}).Error
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

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
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
		AccessKey:    accessKey,
		PriorityMask: uint32(priorityMask),
		ResourceMask: uint32(resourceMask),
	}

	if accessSecret != "" {
		accessAuth.Secret = HashPassword(accessSecret)
		log.Println("access_secret", accessAuth.Secret)
	}

	if err := s.SaveAccessAuth(accessAuth); err != nil {
		return err
	}

	for _, document := range grantPost.Document {
		accessAuthResource := &models.AccessAuthResource{
			AccessKey:  accessKey,
			ResourceId: document,
			Type:       uint8(models.AccessAuthResourceTypeDocument),
		}
		if err := s.CreateAccessAuthResource(accessAuthResource); err != nil {
			return err
		}
	}

	for _, project := range grantPost.Project {
		accessAuthResource := &models.AccessAuthResource{
			AccessKey:  accessKey,
			ResourceId: project,
			Type:       uint8(models.AccessAuthResourceTypeProject),
		}
		if err := s.CreateAccessAuthResource(accessAuthResource); err != nil {
			return err
		}
	}

	for _, team := range grantPost.Team {
		accessAuthResource := &models.AccessAuthResource{
			AccessKey:  accessKey,
			ResourceId: team,
			Type:       uint8(models.AccessAuthResourceTypeTeam),
		}
		if err := s.CreateAccessAuthResource(accessAuthResource); err != nil {
			return err
		}
	}

	return nil
}
