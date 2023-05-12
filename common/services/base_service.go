package services

import (
	"errors"
	"gorm.io/gorm"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/utils/reflect"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
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

type OrderLimitArgs struct {
	Order string
	Limit int
}

type JoinArgs struct {
	Join string
	Args []interface{}
}

type SelectArgs struct {
	Select string
	Args   []interface{}
}

type GroupArgs struct {
	Group string
}

type Unscoped struct{} // 不自动加入软删除字段

/*
参数说明
--参数名-----------类型-----------------------------示例--
modelData：		*models.BaseModel 				&models.User{}
modelDataList：	*[]models.BaseModel 			&[]models.User{}
args：			[>=2]{string, ...interface{}}	"id = ?", 1
args：			[]WhereArgs						返回id=1且type=1或3<=id<=5且type=2的数据：\
[]WhereArgs {\
	WhereArgs{Query:"id = ? and type = ?", Args:[]interface{}{1,1}},\
	WhereArgs{Query:"id >= ? and id <= ? and type = ?", Args:[]interface{}{3,5,2}},\
}
args:			[]OrderLimitArgs				返回按id倒序排序的前10条数据：\
[]OrderLimitArgs{\
	OrderLimitArgs{Order:"id desc", Limit:10},\
}
args:			[]JoinArgs						document left join user：\
[]JoinArgs{\
	JoinArgs{\
		Join:"left join user on document.user_id = user.id",\
		Args:[]interface{}{},\
}
args:			[]SelectArgs					只返回特定字段：\
[]SelectArgs{\
	SelectArgs{\
		Select:"document.*, user.name",\
		Args:[]interface{}{}},\
}
*/

func AddCond(db *gorm.DB, args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}

	othersArgs := args

	_, ok := args[0].(string)
	if ok {
		var i int
		for _, arg := range args[1:] {
			_, ok := arg.(WhereArgs)
			if !ok {
				_, ok = arg.(OrderLimitArgs)
			}
			if !ok {
				_, ok = arg.(JoinArgs)
			}
			if !ok {
				_, ok = arg.(SelectArgs)
			}
			if !ok {
				_, ok = arg.(GroupArgs)
			}
			if !ok {
				_, ok = arg.(Unscoped)
			}
			if ok {
				break
			}
			i++
		}
		db.Where(args[0], args[1:1+i]...)
		othersArgs = args[1+i:]
	}

	firstWhere := false
	for _, arg := range othersArgs {
		switch argv := arg.(type) {
		case WhereArgs:
			if !firstWhere {
				db.Where(argv.Query, argv.Args...)
				firstWhere = true
			} else {
				db.Or(argv.Query, argv.Args...)
			}
		case OrderLimitArgs:
			if argv.Order != "" {
				db.Order(argv.Order)
			}
			if argv.Limit > 0 {
				db.Limit(argv.Limit)
			}
		case JoinArgs:
			if argv.Join != "" {
				db.Joins(argv.Join, argv.Args...)
			}
		case SelectArgs:
			if argv.Select != "" {
				db.Select(argv.Select, argv.Args...)
			}
		case GroupArgs:
			if argv.Group != "" {
				db.Group(argv.Group)
			}
		case Unscoped:
			db.Unscoped()
		default:
			return errors.New("参数格式错误")
		}
	}

	return nil
}

func (s *DefaultService) Create(modelData models.ModelData) error {
	idPtr, ok := reflect.FieldByName(modelData, "Id").(*int64)
	if !ok {
		return errors.New("modelData类型错误")
	}
	*idPtr = snowflake.NextId()
	return s.DB.Model(s.Model).Create(modelData).Error
}

var ErrRecordNotFound = errors.New("记录不存在")

func (s *DefaultService) Get(modelData models.ModelData, args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return err
	}
	if err := db.First(modelData).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else {
		return ErrRecordNotFound
	}
}

func (s *DefaultService) GetById(id int64, modelData models.ModelData) error {
	if err := s.DB.Model(s.Model).Where("id = ?", id).First(modelData).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else {
		return ErrRecordNotFound
	}
}

func (s *DefaultService) Find(modelDataList models.ModelListData, args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return err
	}
	return db.Find(modelDataList).Error
}

// HasWhere 判断是否存在搜索条件
func HasWhere(args ...interface{}) bool {
	// args首个元素为string或args中存在WhereArgs
	return len(args) > 0 && (str.IsString(args[0]) || len(sliceutil.Filter(func(item interface{}) bool {
		_, ok := item.(WhereArgs)
		return ok
	}, args...)) > 0)
}

// AddIdCondIfNoWhere 如果没有传入搜索条件，则添加modelData中的Id字段作为搜索条件
// 若原本就存在搜索条件，或成功添加Id条件，则返回true，否则返回false
func AddIdCondIfNoWhere(modelData models.ModelData, args ...interface{}) bool {
	// 如果没有传入搜索条件，则添加modelData中的Id字段作为搜索条件
	if !HasWhere(args...) {
		// 无传入Id
		if id := modelData.GetId(); id <= 0 {
			return false
		} else {
			args = append(args, WhereArgs{Query: "id = ?", Args: []interface{}{id}})
		}
	}
	return true
}

func (s *DefaultService) Updates(modelData models.ModelData, args ...interface{}) error {
	if !AddIdCondIfNoWhere(modelData, args...) {
		return errors.New("无搜索条件")
	}
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return err
	}
	tx := db.Updates(modelData)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (s *DefaultService) UpdatesById(id int64, modelData models.ModelData) error {
	return s.Updates(modelData, "id = ?", id)
}

func (s *DefaultService) UpdateColumns(values interface{}, args ...interface{}) error {
	if len(args) == 0 {
		return errors.New("无搜索条件")
	}
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return err
	}
	tx := db.UpdateColumns(values)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (s *DefaultService) UpdateColumnsById(id int64, values interface{}) error {
	return s.UpdateColumns(values, "id = ?", id)
}

func (s *DefaultService) Delete(args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if !HasWhere(args...) {
		return errors.New("无搜索条件")
	}
	if err := AddCond(db, args...); err != nil {
		return err
	}
	tx := db.Delete(s.Model)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (s *DefaultService) DeleteById(id int64) error {
	return s.Delete("id = ?", id)
}

// HardDelete 硬删除
func (s *DefaultService) HardDelete(args ...interface{}) error {
	return s.Delete(append(args, Unscoped{}))
}

func (s *DefaultService) HardDeleteById(id int64) error {
	return s.HardDelete("id = ?", id)
}

func (s *DefaultService) Count(count *int64, args ...interface{}) error {
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return err
	}
	return db.Count(count).Error
}
