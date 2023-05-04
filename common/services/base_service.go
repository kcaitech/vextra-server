package services

import (
	"errors"
	"gorm.io/gorm"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/snowflake"
)

type BaseService interface {
}

type DefaultService struct {
	DB    *gorm.DB
	That  BaseService
	Model models.ModelData
}

type WhereArgs struct {
	Query string
	Args  []interface{}
}

/*
参数说明
--参数名-----------类型-------------------------示例--
modelData：		*models.BaseModel 			&models.User{}
modelDataList：	*[]models.BaseModel 		&[]models.User{}
args：			[2]{string, interface{}}	"id = ?", 1
args：			[]WhereArgs					返回id=1且type=1或3<=id<=5且type=2的数据：\
[]WhereArgs {\
	WhereArgs{Query:"id = ? AND type = ?", Args:[]interface{}{1,1}},\
	WhereArgs{Query:"id >= ? AND id <= ? AND type = ?", Args:[]interface{}{3,5,2}},\
}
*/

func addWhereCond(db *gorm.DB, args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}
	_, ok := args[0].(string)
	if ok {
		db.Where(args[0], args[1:]...)
		return nil
	}
	for _, arg := range args {
		if _, ok := arg.(WhereArgs); !ok {
			return errors.New("where查询参数格式错误")
		}
	}
	arg1, _ := args[0].(WhereArgs)
	db.Where(arg1.Query, arg1.Args...)
	for _, arg := range args[1:] {
		_arg, _ := arg.(WhereArgs)
		db.Or(_arg.Query, _arg.Args...)
	}
	return nil
}

func (s *DefaultService) Create(modelData models.ModelData) error {
	data, ok := modelData.(*models.BaseModel)
	if !ok {
		return errors.New("modelData类型错误")
	}
	data.Id = snowflake.NextId()
	return s.DB.Model(s.Model).Create(modelData).Error
}

func (s *DefaultService) Get(modelData models.ModelData, args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := addWhereCond(db, args...); err != nil {
		return err
	}
	return db.First(modelData).Error
}

func (s *DefaultService) GetById(id int64, modelData models.ModelData) error {
	return s.DB.Model(s.Model).Where("id = ?", id).First(modelData).Error
}

func (s *DefaultService) Find(modelDataList models.ModelListData, order string, limit int, args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := addWhereCond(db, args...); err != nil {
		return err
	}
	return db.Find(modelDataList).Order(order).Limit(limit).Error
}

func (s *DefaultService) Updates(modelData models.ModelData, args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := addWhereCond(db, args...); err != nil {
		return err
	}
	return db.Updates(modelData).Error
}

func (s *DefaultService) UpdatesById(id int64, modelData models.ModelData) error {
	return s.DB.Model(s.Model).Where("id = ?", id).Updates(modelData).Error
}

func (s *DefaultService) Delete(args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := addWhereCond(db, args...); err != nil {
		return err
	}
	return db.Delete(s.Model).Error
}

func (s *DefaultService) DeleteById(id int64) error {
	return s.DB.Model(s.Model).Where("id = ?", id).Delete(s.Model).Error
}
