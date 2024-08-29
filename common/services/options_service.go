package services

import (
	"errors"
	"protodesign.cn/kcserver/common/models"
)

type OptionsService struct {
	*DefaultService
}

func NewOptionsService() *OptionsService {
	that := &OptionsService{
		DefaultService: NewDefaultService(&models.Options{}),
	}
	that.That = that
	return that
}

func (s *OptionsService) GetOne(_type string) (string, error) {
	var result models.Options
	if s.Get(
		&result,
		"deleted_at is null and type = ?",
		_type,
		&OrderLimitArgs{"id desc", 0},
	) == nil {
		return result.Detail, nil
	}
	return "", errors.New("not found")
}

func (s *OptionsService) SetOne(_type, detail string) bool {
	var result models.Options
	if s.Get(
		&result,
		"deleted_at is null and type = ?",
		_type,
		&OrderLimitArgs{"id desc", 0},
	) == nil {
		result.Detail = detail
		count, _ := s.Updates(&result)
		return count > 0
	}

	result.Type = _type
	result.Detail = detail
	return s.Create(&result) == nil
}
