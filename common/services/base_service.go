package services

import (
	"errors"
	"fmt"
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

func MergeWhereArgs(connector string, whereArgs []*WhereArgs) *WhereArgs {
	var query string
	var args []interface{}
	for _, whereArg := range whereArgs {
		if whereArg.Query == "" {
			continue
		}
		if query == "" {
			query = fmt.Sprintf("(%s)", whereArg.Query)
		} else {
			query = fmt.Sprintf("((%s) %s (%s))", query, connector, whereArg.Query)
		}
		args = append(args, whereArg.Args...)
	}
	return &WhereArgs{Query: query, Args: args}
}

func MergeWhereArgsAnd(whereArgs []*WhereArgs) *WhereArgs {
	return MergeWhereArgs("and", whereArgs)
}

func MergeWhereArgsOr(whereArgs []*WhereArgs) *WhereArgs {
	return MergeWhereArgs("or", whereArgs)
}

func NotWhereArgs(whereArgs *WhereArgs) *WhereArgs {
	return &WhereArgs{Query: fmt.Sprintf("(not (%s))", whereArgs.Query), Args: whereArgs.Args}
}

type WhereNodeType uint8

const (
	WhereNodeTypeAnd WhereNodeType = iota
	WhereNodeTypeOr
	WhereNodeTypeNot
	WhereNodeTypeValue
)

// WhereNode Where逻辑树节点
type WhereNode struct {
	Type      WhereNodeType
	Left      *WhereNode
	Right     *WhereNode
	WhereArgs *WhereArgs
	Error     error
}

func joinErrors(err0 error, err1 error) error {
	var err error
	if err0 != nil || err1 != nil {
		if err0 != nil && err1 != nil {
			err = errors.Join(err0, err1)
		} else if err0 != nil {
			err = err0
		} else {
			err = err1
		}
	}
	return err
}

func (n *WhereNode) Calc() (*WhereArgs, error) {
	valL, errL := n.Left.Calc()
	valR, errR := n.Right.Calc()
	err := joinErrors(errL, errR)
	switch n.Type {
	case WhereNodeTypeAnd:
		if err != nil {
			return &WhereArgs{}, err
		}
		return MergeWhereArgsAnd([]*WhereArgs{valL, valR}), nil
	case WhereNodeTypeOr:
		if err != nil {
			return &WhereArgs{}, err
		}
		return MergeWhereArgsOr([]*WhereArgs{valL, valR}), nil
	case WhereNodeTypeNot:
		if errL != nil {
			return &WhereArgs{}, errL
		}
		return NotWhereArgs(valL), nil
	case WhereNodeTypeValue:
		return n.WhereArgs, nil
	}
	return &WhereArgs{}, errors.New("不支持的WhereNodeType")
}

func ConvertToWhereNode(val any) (*WhereNode, error) {
	switch v := val.(type) {
	case *WhereNode:
		return v, nil
	case WhereNode:
		return &v, nil
	case *WhereArgs:
		return &WhereNode{Type: WhereNodeTypeValue, WhereArgs: v}, nil
	case WhereArgs:
		return &WhereNode{Type: WhereNodeTypeValue, WhereArgs: &v}, nil
	}
	return nil, errors.New("参数错误：仅支持*WhereNode, WhereNode, *WhereArgs, WhereArgs类型")
}

func WhereNodeAnd(node0 any, node1 any) *WhereNode {
	left, errL := ConvertToWhereNode(node0)
	right, errR := ConvertToWhereNode(node1)
	err := joinErrors(errL, errR)
	return &WhereNode{Type: WhereNodeTypeAnd, Left: left, Right: right, Error: err}
}

func WhereNodeOr(node0 any, node1 any) *WhereNode {
	left, errL := ConvertToWhereNode(node0)
	right, errR := ConvertToWhereNode(node1)
	err := joinErrors(errL, errR)
	return &WhereNode{Type: WhereNodeTypeOr, Left: left, Right: right, Error: err}
}

func WhereNodeNot(node any) *WhereNode {
	left, err := ConvertToWhereNode(node)
	return &WhereNode{Type: WhereNodeTypeNot, Left: left, Error: err}
}

func (n *WhereNode) And(node any) *WhereNode {
	return WhereNodeAnd(n, node)
}

func (n *WhereNode) Or(node any) *WhereNode {
	return WhereNodeOr(n, node)
}

func (n *WhereNode) Not() *WhereNode {
	return WhereNodeNot(n)
}

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
				_, ok = arg.(*WhereArgs)
			}
			if !ok {
				_, ok = arg.([]WhereArgs)
			}
			if !ok {
				_, ok = arg.([]*WhereArgs)
			}
			if !ok {
				_, ok = arg.(WhereNode)
			}
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
	addWhereCond := func(query string, args ...interface{}) {
		if !firstWhere {
			db.Where(query, args...)
			firstWhere = true
		} else {
			db.Or(query, args...)
		}
	}

	for _, arg := range othersArgs {
		switch argv := arg.(type) {
		case WhereArgs:
			addWhereCond(argv.Query, argv.Args...)
		case *WhereArgs:
			addWhereCond(argv.Query, argv.Args...)
		case []WhereArgs:
			whereArgs := MergeWhereArgsAnd(
				sliceutil.MapT(func(item WhereArgs) *WhereArgs {
					return &item
				}, argv...),
			)
			addWhereCond(whereArgs.Query, whereArgs.Args...)
		case []*WhereArgs:
			whereArgs := MergeWhereArgsAnd(argv)
			addWhereCond(whereArgs.Query, whereArgs.Args...)
		case WhereNode:
			if argv.Error != nil {
				return argv.Error
			}
			whereArgs, err := argv.Calc()
			if err != nil {
				return err
			}
			addWhereCond(whereArgs.Query, whereArgs.Args...)
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
