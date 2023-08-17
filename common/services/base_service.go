package services

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/snowflake"
	myReflect "protodesign.cn/kcserver/utils/reflect"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type BaseService interface {
	GetTable() string
}

type DefaultService struct {
	DB              *gorm.DB
	That            BaseService
	Model           models.ModelData
	Table           string
	FieldNames      []string
	TableFieldNames []string
}

func (s *DefaultService) GetTable() string {
	return s.Table
}

func (s *DefaultService) GetFieldNames() []string {
	return s.FieldNames
}

func (s *DefaultService) GetTableFieldNames() []string {
	return s.TableFieldNames
}

func (s *DefaultService) GetFieldNamesStr() string {
	return strings.Join(s.FieldNames, ",")
}

func (s *DefaultService) GetTableFieldNamesStr() string {
	return strings.Join(s.TableFieldNames, ",")
}

func (s *DefaultService) GetTableFieldNamesStrAliasByPrefix(prefix string) string {
	var tableFieldNames []string
	for _, tableFieldName := range s.FieldNames {
		tableFieldNames = append(tableFieldNames, fmt.Sprintf("%s.%s as %s", s.Table, tableFieldName, prefix+tableFieldName))
	}
	return strings.Join(tableFieldNames, ",")
}

func (s *DefaultService) GetTableFieldNamesStrAliasByDefaultPrefix(connector string) string {
	if connector == "" {
		connector = "__"
	}
	return s.GetTableFieldNamesStrAliasByPrefix(s.Table + connector)
}

func NewDefaultService(model models.ModelData) *DefaultService {
	s := DefaultService{
		DB:    models.DB,
		Model: model,
	}
	db := models.DB.Model(model)
	_ = db.Statement.Parse(&model)
	s.Table = db.Statement.Table
	for _, field := range db.Statement.Schema.Fields {
		s.FieldNames = append(s.FieldNames, field.DBName)
		s.TableFieldNames = append(s.TableFieldNames, fmt.Sprintf("%s.%s", s.Table, field.DBName))
	}
	return &s
}

type WhereArgs struct {
	Query string
	Args  []any
}

type OrderLimitArgs struct {
	Order string
	Limit int
}

type JoinArgsRaw struct {
	Join string
	Args []any
}

type JoinType string

const (
	JoinTypeInner JoinType = "inner join"
	JoinTypeLeft  JoinType = "left join"
	JoinTypeRight JoinType = "right join"
)

type JoinArgsOn struct {
	Field     string
	JoinField string
	JoinTable string
}

type JoinArgs struct {
	TableName         string
	OriginalTableName string
	Type              JoinType
	On                []JoinArgsOn
	OnArgs            []any
}

type SelectArgs struct {
	Select string
	Args   []any
}

type GroupArgs struct {
	Group string
}

type Unscoped struct{} // 不自动加入软删除字段

// As 指定表的别名
type As struct {
	BaseService BaseService
	Alias       string
}

// Wrap 对多个参数的包装
type Wrap struct {
	Args []any
}

func MergeWhereArgs(connector string, whereArgs []*WhereArgs) *WhereArgs {
	var query string
	var args []any
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

func MergeSelectArgs(selectArgs []*SelectArgs) *SelectArgs {
	var _select string
	var args []any
	for _, selectArg := range selectArgs {
		if selectArg.Select == "" {
			continue
		}
		if _select == "" {
			_select = selectArg.Select
		} else {
			_select += fmt.Sprintf(",%s", selectArg.Select)
		}
		args = append(args, selectArg.Args...)
	}
	return &SelectArgs{Select: _select, Args: args}
}

/*
参数说明
--参数名-----------类型-----------------------------示例--
modelData：		*models.BaseModel 				&models.User{}
modelDataList：	*[]models.BaseModel 			&[]models.User{}
args：			[>=2]{string, ...interface{}}	"id = ?", 1
args：			[]WhereArgs						返回id=1且type=1或3<=id<=5且type=2的数据：\
[]WhereArgs {\
	WhereArgs{Query:"id = ? and type = ?", Args:[]any{1,1}},\
	WhereArgs{Query:"id >= ? and id <= ? and type = ?", Args:[]any{3,5,2}},\
}
args:			[]OrderLimitArgs				返回按id倒序排序的前10条数据：\
[]OrderLimitArgs{\
	OrderLimitArgs{Order:"id desc", Limit:10},\
}
args:			[]JoinArgs						document left join user：\
[]JoinArgs{\
	JoinArgs{\
		Join:"left join user on document.user_id = user.id",\
		Args:[]any{},\
}
args:			[]SelectArgs					只返回特定字段：\
[]SelectArgs{\
	SelectArgs{\
		Select:"document.*, user.name",\
		Args:[]any{}},\
}
*/

func AddCond(db *gorm.DB, args ...any) error {
	if len(args) == 0 {
		return nil
	}

	othersArgs := args

	_, ok := args[0].(string)
	if ok {
		var i int
	Loop:
		for _, arg := range args[1:] {
			switch arg.(type) {
			case WhereArgs, *WhereArgs, []WhereArgs, []*WhereArgs, *[]WhereArgs, *[]*WhereArgs,
				WhereNode, *WhereNode,
				OrderLimitArgs, *OrderLimitArgs,
				JoinArgsRaw, *JoinArgsRaw,
				JoinArgs, *JoinArgs, []JoinArgs, []*JoinArgs, *[]JoinArgs, *[]*JoinArgs,
				SelectArgs, *SelectArgs, []SelectArgs, *[]SelectArgs, []*SelectArgs, *[]*SelectArgs,
				GroupArgs, *GroupArgs,
				Unscoped, *Unscoped,
				As, *As,
				Wrap, *Wrap:
				break Loop
			}
			i++
		}
		db.Where(args[0], args[1:1+i]...)
		othersArgs = args[1+i:]
	}

	firstWhere := false
	addWhereCond := func(query string, args ...any) {
		if !firstWhere {
			db.Where(query, args...)
			firstWhere = true
		} else {
			db.Or(query, args...)
		}
	}

	selectArgsList := make([]*SelectArgs, 0)

	calcWhereArgs := func(arg any) error {
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
		case *[]WhereArgs:
			whereArgs := MergeWhereArgsAnd(
				sliceutil.MapT(func(item WhereArgs) *WhereArgs {
					return &item
				}, *argv...),
			)
			addWhereCond(whereArgs.Query, whereArgs.Args...)
		case *[]*WhereArgs:
			whereArgs := MergeWhereArgsAnd(*argv)
			addWhereCond(whereArgs.Query, whereArgs.Args...)
		}
		return nil
	}

	calcWhereNode := func(arg any) error {
		var argv WhereNode
		var ok bool
		if argv, ok = arg.(WhereNode); !ok {
			if argv1, ok := arg.(*WhereNode); !ok {
				return nil
			} else {
				argv = *argv1
			}
		}
		if argv.Error != nil {
			return argv.Error
		}
		whereArgs, err := argv.Calc()
		if err != nil {
			return err
		}
		addWhereCond(whereArgs.Query, whereArgs.Args...)
		return nil
	}

	calcOrderLimit := func(arg any) error {
		var argv OrderLimitArgs
		var ok bool
		if argv, ok = arg.(OrderLimitArgs); !ok {
			if argv1, ok := arg.(*OrderLimitArgs); !ok {
				return nil
			} else {
				argv = *argv1
			}
		}
		if argv.Order != "" {
			db.Order(argv.Order)
		}
		if argv.Limit > 0 {
			db.Limit(argv.Limit)
		}
		return nil
	}

	calcJoinArgsRaw := func(arg any) error {
		var argv JoinArgsRaw
		var ok bool
		if argv, ok = arg.(JoinArgsRaw); !ok {
			if argv1, ok := arg.(*JoinArgsRaw); !ok {
				return nil
			} else {
				argv = *argv1
			}
		}
		if argv.Join != "" {
			db.Joins(argv.Join, argv.Args...)
		}
		return nil
	}

	_calcJoinArgs := func(arg any) error {
		var argv JoinArgs
		var ok bool
		if argv, ok = arg.(JoinArgs); !ok {
			if argv1, ok := arg.(*JoinArgs); !ok {
				return nil
			} else {
				argv = *argv1
			}
		}
		if argv.TableName == "" || argv.OriginalTableName == "" {
			return errors.New("JoinArgs.Table为空")
		}
		if argv.Type != JoinTypeInner && argv.Type != JoinTypeLeft && argv.Type != JoinTypeRight {
			return errors.New("JoinArgs.Type错误")
		}
		if argv.On == nil {
			return errors.New("JoinArgs.On为空")
		}
		argsNum := 0
		onList := make([]string, 0, len(argv.On))
		for _, on := range argv.On {
			if on.Field == "" || on.JoinTable == "" || on.JoinField == "" {
				return errors.New("JoinArgs.On.Field、JoinArgs.On.JoinTable、JoinArgs.On.JoinField不可为空")
			}
			if on.JoinField == "?" {
				argsNum++
				if on.JoinTable != "" {
					on.JoinTable = ""
				}
			} else {
				on.JoinTable += "."
			}
			onList = append(onList, fmt.Sprintf("%s.%s=%s%s", argv.TableName, on.Field, on.JoinTable, on.JoinField))
		}
		if len(argv.OnArgs) != argsNum {
			return errors.New("JoinArgs.OnArgs错误")
		}
		// join table别名
		tableAlias := ""
		if argv.OriginalTableName != argv.TableName {
			tableAlias = " " + argv.TableName
		}
		db.Joins(fmt.Sprintf("%s %s%s on %s", argv.Type, argv.OriginalTableName, tableAlias, strings.Join(onList, " and ")), argv.OnArgs...)
		return nil
	}

	calcJoinArgs := func(arg any) error {
		switch argv := arg.(type) {
		case JoinArgs, *JoinArgs:
			if err := _calcJoinArgs(argv); err != nil {
				return err
			}
		case []JoinArgs:
			for _, joinArgs := range argv {
				if err := _calcJoinArgs(joinArgs); err != nil {
					return err
				}
			}
		case []*JoinArgs:
			for _, joinArgs := range argv {
				if err := _calcJoinArgs(joinArgs); err != nil {
					return err
				}
			}
		case *[]JoinArgs:
			for _, joinArgs := range *argv {
				if err := _calcJoinArgs(joinArgs); err != nil {
					return err
				}
			}
		case *[]*JoinArgs:
			for _, joinArgs := range *argv {
				if err := _calcJoinArgs(joinArgs); err != nil {
					return err
				}
			}
		}
		return nil
	}

	calcSelectArgs := func(arg any) error {
		switch argv := arg.(type) {
		case SelectArgs:
			if argv.Select != "" {
				selectArgsList = append(selectArgsList, &argv)
			}
		case *SelectArgs:
			if argv.Select != "" {
				selectArgsList = append(selectArgsList, argv)
			}
		case []SelectArgs:
			for _, selectArgs := range argv {
				if selectArgs.Select != "" {
					selectArgsList = append(selectArgsList, &selectArgs)
				}
			}
		case []*SelectArgs:
			for _, selectArgs := range argv {
				if selectArgs.Select != "" {
					selectArgsList = append(selectArgsList, selectArgs)
				}
			}
		case *[]SelectArgs:
			for _, selectArgs := range *argv {
				if selectArgs.Select != "" {
					selectArgsList = append(selectArgsList, &selectArgs)
				}
			}
		case *[]*SelectArgs:
			for _, selectArgs := range *argv {
				if selectArgs.Select != "" {
					selectArgsList = append(selectArgsList, selectArgs)
				}
			}
		}
		return nil
	}

	calcGroupArgs := func(arg any) error {
		switch argv := arg.(type) {
		case GroupArgs:
			if argv.Group != "" {
				db.Group(argv.Group)
			}
		case *GroupArgs:
			if argv.Group != "" {
				db.Group(argv.Group)
			}
		}
		return nil
	}

	calcUnscoped := func(arg any) error {
		switch arg.(type) {
		case Unscoped, *Unscoped:
			db.Unscoped()
		}
		return nil
	}

	calcAs := func(arg any) error {
		switch argv := arg.(type) {
		case As:
			if argv.BaseService != nil && argv.Alias != "" {
				db.Table(fmt.Sprintf("%s as %s", argv.BaseService.GetTable(), argv.Alias))
			}
		case *As:
			if argv.BaseService != nil && argv.Alias != "" {
				db.Table(fmt.Sprintf("%s as %s", argv.BaseService.GetTable(), argv.Alias))
			}
		}
		return nil
	}

	var calc func(args []any) error
	calc = func(args []any) error {
		for _, arg := range args {
			switch argv := arg.(type) {
			case WhereArgs, *WhereArgs, []WhereArgs, []*WhereArgs, *[]WhereArgs, *[]*WhereArgs:
				if err := calcWhereArgs(argv); err != nil {
					return err
				}
			case WhereNode, *WhereNode:
				if err := calcWhereNode(argv); err != nil {
					return err
				}
			case OrderLimitArgs, *OrderLimitArgs:
				if err := calcOrderLimit(argv); err != nil {
					return err
				}
			case JoinArgsRaw, *JoinArgsRaw:
				if err := calcJoinArgsRaw(argv); err != nil {
					return err
				}
			case JoinArgs, *JoinArgs, []JoinArgs, []*JoinArgs, *[]JoinArgs, *[]*JoinArgs:
				if err := calcJoinArgs(argv); err != nil {
					return err
				}
			case SelectArgs, *SelectArgs, []SelectArgs, []*SelectArgs, *[]SelectArgs, *[]*SelectArgs:
				if err := calcSelectArgs(argv); err != nil {
					return err
				}
			case GroupArgs, *GroupArgs:
				if err := calcGroupArgs(argv); err != nil {
					return err
				}
			case Unscoped, *Unscoped:
				if err := calcUnscoped(argv); err != nil {
					return err
				}
			case As, *As:
				if err := calcAs(argv); err != nil {
					return err
				}
			case Wrap:
				if argv.Args != nil {
					if err := calc(argv.Args); err != nil {
						return err
					}
				}
			case *Wrap:
				if argv.Args != nil {
					if err := calc(argv.Args); err != nil {
						return err
					}
				}
			default:
				return errors.New("参数格式错误")
			}
		}
		return nil
	}

	if err := calc(othersArgs); err != nil {
		return err
	}

	if len(selectArgsList) > 0 {
		allSelectArgs := MergeSelectArgs(selectArgsList)
		if allSelectArgs.Select != "" {
			db.Select(allSelectArgs.Select, allSelectArgs.Args...)
		}
	}

	return nil
}

func (s *DefaultService) Create(modelData models.ModelData) error {
	modelData.SetId(snowflake.NextId())
	return s.DB.Model(s.Model).Create(modelData).Error
}

var ErrRecordNotFound = errors.New("记录不存在")

func (s *DefaultService) Get(modelData models.ModelData, args ...any) error {
	db := s.DB.Model(s.Model)
	paramArgs := GetParamArgsFromArgs(&args)
	args = append(args, GenerateJoinArgs(modelData, s.Table, paramArgs))
	args = append(args, GenerateSelectArgs(modelData, ""))
	if err := AddCond(db, args...); err != nil {
		return err
	}
	if err := db.First(modelData).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	} else {
		return err
	}
}

func (s *DefaultService) GetById(id int64, modelData models.ModelData) error {
	if err := s.DB.Model(s.Model).Where("id = ?", id).First(modelData).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else {
		return ErrRecordNotFound
	}
}

func (s *DefaultService) Find(modelDataList models.ModelListData, args ...any) error {
	db := s.DB.Model(s.Model)
	paramArgs := GetParamArgsFromArgs(&args)
	args = append(args, GenerateJoinArgs(modelDataList, s.Table, paramArgs))
	args = append(args, GenerateSelectArgs(modelDataList, ""))
	if err := AddCond(db, args...); err != nil {
		return err
	}
	return db.Find(modelDataList).Error
}

// HasWhere 判断是否存在搜索条件
func HasWhere(args *[]any) bool {
	// args首个元素为string或args中存在WhereArgs
	return len(*args) > 0 && (str.IsString((*args)[0]) || len(sliceutil.Filter(func(item any) bool {
		_, ok := item.(WhereArgs)
		return ok
	}, *args...)) > 0)
}

// AddIdCondIfNoWhere 如果没有传入搜索条件，则添加modelData中的Id字段作为搜索条件
// 若原本就存在搜索条件，或成功添加Id条件，则返回true，否则返回false
func AddIdCondIfNoWhere(modelData models.ModelData, args *[]any) bool {
	// 如果没有传入搜索条件，则添加modelData中的Id字段作为搜索条件
	if !HasWhere(args) {
		// 无传入Id
		if id := modelData.GetId(); id <= 0 {
			return false
		} else {
			*args = append(*args, WhereArgs{Query: "id = ?", Args: []any{id}})
		}
	}
	return true
}

func (s *DefaultService) updates(modelData models.ModelData, ignoreZero bool, args ...any) (int64, error) {
	if !AddIdCondIfNoWhere(modelData, &args) {
		return 0, errors.New("无搜索条件")
	}
	db := s.DB.Model(s.Model)
	paramArgs := GetParamArgsFromArgs(&args)
	args = append(args, GenerateJoinArgs(modelData, s.Table, paramArgs))
	args = append(args, GenerateSelectArgs(modelData, ""))
	if !ignoreZero {
		db.Select("*")
	}
	if err := AddCond(db, args...); err != nil {
		return 0, err
	}
	tx := db.Updates(modelData)
	if tx.Error != nil {
		return 0, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 0, ErrRecordNotFound
	}
	return tx.RowsAffected, nil
}

func (s *DefaultService) Updates(modelData models.ModelData, args ...any) (int64, error) {
	return s.updates(modelData, false, args...)
}

func (s *DefaultService) UpdatesById(id int64, modelData models.ModelData) (int64, error) {
	return s.Updates(modelData, "id = ?", id)
}

func (s *DefaultService) UpdatesIgnoreZero(modelData models.ModelData, args ...any) (int64, error) {
	return s.updates(modelData, true, args...)
}

func (s *DefaultService) UpdatesIgnoreZeroById(id int64, modelData models.ModelData) (int64, error) {
	return s.UpdatesIgnoreZero(modelData, "id = ?", id)
}

func Expr(expr string, args ...interface{}) clause.Expr {
	return gorm.Expr(expr, args...)
}

func (s *DefaultService) UpdateColumns(values map[string]any, args ...any) (int64, error) {
	if len(args) == 0 {
		return 0, errors.New("无搜索条件")
	}
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return 0, err
	}
	values["updated_at"] = myTime.Time(time.Now())
	tx := db.UpdateColumns(values)
	if tx.Error != nil {
		return 0, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 0, ErrRecordNotFound
	}
	return tx.RowsAffected, nil
}

func (s *DefaultService) UpdateColumnsById(id int64, values map[string]any) (int64, error) {
	return s.UpdateColumns(values, "id = ?", id)
}

// Delete 软删除
func (s *DefaultService) Delete(args ...any) (int64, error) {
	db := s.DB.Model(s.Model)
	if !HasWhere(&args) {
		return 0, errors.New("无搜索条件")
	}
	if err := AddCond(db, args...); err != nil {
		return 0, err
	}
	tx := db.Delete(s.Model)
	if tx.Error != nil {
		return 0, tx.Error
	}
	if tx.RowsAffected == 0 {
		return 0, ErrRecordNotFound
	}
	return tx.RowsAffected, nil
}

func (s *DefaultService) DeleteById(id int64) (int64, error) {
	return s.Delete("id = ?", id)
}

// HardDelete 硬删除
func (s *DefaultService) HardDelete(args ...any) (int64, error) {
	return s.Delete(append(args, Unscoped{})...)
}

func (s *DefaultService) HardDeleteById(id int64) (int64, error) {
	return s.HardDelete("id = ?", id)
}

func (s *DefaultService) Count(count *int64, args ...any) error {
	db := s.DB.Model(s.Model)
	if err := AddCond(db, args...); err != nil {
		return err
	}
	return db.Count(count).Error
}

func (s *DefaultService) Exist(args ...any) (bool, error) {
	var count int64
	if err := s.Count(&count, args...); err != nil {
		return false, err
	}
	return count > 0, nil
}

// 迭代dataType的字段，匿名字段递归调用，非匿名struct字段调用回调处理函数
func iterateDataTypeFields(dataType reflect.Type, handler func(typeField reflect.StructField, typeFieldType reflect.Type)) {
	modelDataType := myReflect.EnterPointer(dataType)
	if modelDataType.Kind() == reflect.Slice {
		modelDataType = modelDataType.Elem()
	}
	if modelDataType.Kind() != reflect.Struct {
		return
	}
	for i, num := 0, modelDataType.NumField(); i < num; i++ {
		typeField := modelDataType.Field(i)
		if !typeField.IsExported() {
			continue
		}
		typeFieldType := myReflect.EnterPointer(typeField.Type)
		if typeFieldType.Kind() != reflect.Struct {
			continue
		}
		// 匿名字段递归调用
		if typeField.Anonymous || typeField.Tag.Get("anonymous") == "true" {
			iterateDataTypeFields(typeField.Type, handler)
			continue
		}
		handler(typeField, typeFieldType)
	}
}

// 根据dataType生成SelectArgs
// 迭代dataType的字段：
// 匿名字段递归调用
// 非匿名字段且带有table标签的struct类型字段：获取其内部的所有非匿名公开字段，对于其内部的匿名公开struct字段则递归获取
func generateSelectArgs(dataType reflect.Type, connector string, selectArgsList *[]*SelectArgs) {
	if connector == "" {
		connector = "__"
	}
	iterateDataTypeFields(dataType, func(typeField reflect.StructField, typeFieldType reflect.Type) {
		tableName, ok := typeField.Tag.Lookup("table")
		if !ok { // 无table标签时取join标签的第一个参数
			join, ok1 := typeField.Tag.Lookup("join")
			if !ok1 { // 无join标签，跳过
				return
			}
			splitRes := strings.Split(join, ";")
			if len(splitRes) > 0 {
				tableName = splitRes[0]
			}
		}
		splitRes := strings.Split(tableName, ",")
		if tableName == "" || len(splitRes) == 0 { // 未指定tableName，自动根据字段名生成蛇形命名
			tableName = str.CamelToSnake(typeField.Name)
		} else if len(splitRes) == 1 { // 指定tableName
			tableName = splitRes[0]
		} else { // 取别名
			tableName = splitRes[1]
		}
		// 非匿名struct且具有table标签，获取内部所有公开字段
		var GetExportedFieldNames func(fieldType reflect.Type, fieldNames *[]string)
		GetExportedFieldNames = func(fieldType reflect.Type, fieldNames *[]string) {
			for i, num := 0, fieldType.NumField(); i < num; i++ {
				fieldTypeField := fieldType.Field(i)
				if !fieldTypeField.IsExported() {
					continue
				}
				// 匿名而且是结构体或结构体指针的字段
				if fieldTypeField.Anonymous && myReflect.EnterPointer(fieldTypeField.Type).Kind() == reflect.Struct {
					GetExportedFieldNames(fieldTypeField.Type, fieldNames)
					continue
				}
				var name string
				if columnName := fieldTypeField.Tag.Get("column"); columnName != "" {
					name = columnName
				} else {
					name = str.CamelToSnake(fieldTypeField.Name)
				}
				*fieldNames = append(*fieldNames, name)
			}
		}
		fieldNames := make([]string, 0, typeFieldType.NumField())
		GetExportedFieldNames(typeFieldType, &fieldNames)
		if len(fieldNames) == 0 {
			return
		}
		// 生成selectArgs
		fieldAliasNames := make([]string, 0, len(fieldNames))
		for _, fieldName := range fieldNames {
			fieldAliasNames = append(fieldAliasNames, fmt.Sprintf("%s.%s as %s", tableName, fieldName, tableName+connector+fieldName))
		}
		*selectArgsList = append(*selectArgsList, &SelectArgs{strings.Join(fieldAliasNames, ","), nil})
	})
}

// GenerateSelectArgs 根据modelData生成SelectArgs
// table标签决定SQL语句的select部分中是否包含对应的字段，不设置table时会读取join标签的第一个参数作为table，为空字符串时自动根据字段名生成蛇形命名
func GenerateSelectArgs(modelData any, connector string) *[]*SelectArgs {
	selectArgsList := make([]*SelectArgs, 0)
	generateSelectArgs(reflect.TypeOf(modelData), connector, &selectArgsList)
	return &selectArgsList
}

// 根据dataType生成JoinArgs
// 迭代dataType的字段
// 匿名字段递归调用
// 非匿名字段且带有join标签的struct类型字段：解析join标签，生成JoinArgs
func generateJoinArgs(dataType reflect.Type, mainTable string, paramArgs ParamArgs, joinArgsList *[]*JoinArgs) {
	iterateDataTypeFields(dataType, func(typeField reflect.StructField, typeFieldType reflect.Type) {
		join, ok := typeField.Tag.Lookup("join")
		if !ok { // 无join标签，跳过
			return
		}
		// 非匿名struct且具有join标签，解析join标签，生成JoinArgs
		splitRes := strings.Split(join, ";") // 表名,表别名;连接方式;selfField_n,joinField_n;...
		if len(splitRes) < 3 {
			return
		}
		// 取tableName
		tableName := splitRes[0]
		splitRes1 := strings.Split(tableName, ",")
		var originalTableName string
		if tableName == "" || len(splitRes1) == 0 { // 未指定tableName，自动根据字段名生成蛇形命名
			originalTableName = str.CamelToSnake(typeField.Name)
			tableName = originalTableName
		} else if len(splitRes1) == 1 { // 指定tableName
			originalTableName = splitRes1[0]
			tableName = originalTableName
		} else { // 取别名
			originalTableName = splitRes1[0]
			tableName = splitRes1[1]
		}
		// 取joinType
		joinType := JoinType(splitRes[1])
		if joinType == "" {
			return
		}
		if originalTableName == mainTable || tableName == mainTable {
			return
		}
		switch joinType {
		case "inner":
			joinType = JoinTypeInner
		case "left":
			joinType = JoinTypeLeft
		case "right":
			joinType = JoinTypeRight
		default:
			return
		}
		joinTableFieldHandler := func(value string) (joinField string, joinTable string, joinArgs any, res bool) {
			joinTableFieldSplitRes := strings.Split(value, ".")
			if len(joinTableFieldSplitRes) == 1 {
				joinField = joinTableFieldSplitRes[0]
			} else if len(joinTableFieldSplitRes) >= 2 {
				joinTable = joinTableFieldSplitRes[0]
				joinField = joinTableFieldSplitRes[1]
			}
			res = false
			if strings.HasPrefix(joinField, "?") {
				joinFieldParamArgAny, ok := paramArgs[joinField]
				if ok {
					joinField = "?"
					joinTable = "?"
					joinArgs = joinFieldParamArgAny
					res = true
				}
				return
			}
			if strings.HasPrefix(joinField, "#") {
				joinFieldParamArgAny, ok := paramArgs[joinField]
				if joinFieldParamArg, ok1 := joinFieldParamArgAny.(string); ok && ok1 {
					joinField = joinFieldParamArg
				} else {
					return
				}
			}
			if strings.HasPrefix(joinTable, "#") {
				joinTableParamArgAny, ok := paramArgs[joinTable]
				if joinTableParamArg, ok1 := joinTableParamArgAny.(string); ok && ok1 {
					joinTable = joinTableParamArg
				} else {
					return
				}
			}
			res = true
			return
		}
		joinOnArgs := make([]any, 0)
		onList := make([]JoinArgsOn, 0, len(splitRes)-2)
		for _, onStr := range splitRes[2:] {
			joinArgsOn := JoinArgsOn{}
			onSplitRes := strings.Split(onStr, ",")
			if joinArgsOn.Field = onSplitRes[0]; joinArgsOn.Field == "" {
				continue
			}
			if len(onSplitRes) >= 2 {
				matches := regexp.MustCompile(`\[(.*?)]`).FindAllStringSubmatch(onSplitRes[1], -1)
				ok := false
				var joinOnArg any
				if len(matches) > 0 {
					matchRes := matches[0][1]
					matchResSplitRes := strings.Split(matchRes, " ")
					for _, matchResItem := range matchResSplitRes {
						var joinField, joinTable string
						joinField, joinTable, joinOnArg, ok = joinTableFieldHandler(matchResItem)
						if !ok {
							continue
						}
						joinArgsOn.JoinField = joinField
						joinArgsOn.JoinTable = joinTable
						break
					}
				}
				if !ok {
					joinArgsOn.JoinField, joinArgsOn.JoinTable, joinOnArg, ok = joinTableFieldHandler(onSplitRes[1])
				}
				if joinOnArg != nil {
					joinOnArgs = append(joinOnArgs, joinOnArg)
				}
			}
			if joinArgsOn.JoinField == "" {
				joinArgsOn.JoinField = joinArgsOn.Field
			}
			joinTableFieldSplitRes := strings.Split(joinArgsOn.JoinField, ".")
			if len(joinTableFieldSplitRes) >= 2 {
				joinArgsOn.JoinTable = joinTableFieldSplitRes[0]
				joinArgsOn.JoinField = joinTableFieldSplitRes[1]
			}
			if joinArgsOn.JoinTable == "" {
				joinArgsOn.JoinTable = mainTable
			}
			if joinArgsOn.JoinTable == "" {
				continue
			}
			onList = append(onList, joinArgsOn)
		}
		if len(onList) == 0 {
			return
		}
		*joinArgsList = append(*joinArgsList, &JoinArgs{TableName: tableName, OriginalTableName: originalTableName, Type: joinType, On: onList, OnArgs: joinOnArgs})
	})
}

// GenerateJoinArgs 根据modelData生成JoinArgs
// join标签决定SQL语句的join部分中是否包含此表，格式为：表名,表别名;连接方式;selfField_n,joinField_n;...
// 不设置table（第一个参数）时会读取join标签的第一个参数作为table，为空字符串时自动根据字段名生成蛇形命名
func GenerateJoinArgs(modelData any, mainTable string, paramArgs ParamArgs) *[]*JoinArgs {
	joinArgsList := make([]*JoinArgs, 0)
	generateJoinArgs(reflect.TypeOf(modelData), mainTable, paramArgs, &joinArgsList)
	return &joinArgsList
}

type ParamArgs map[string]any

// MergeParamArgs 合并多个ParamArgs
func (that ParamArgs) MergeParamArgs(paramArgsList ...ParamArgs) ParamArgs {
	if len(paramArgsList) == 0 {
		return that
	}
	for _, paramArgs := range paramArgsList {
		for k, v := range paramArgs {
			that[k] = v
		}
	}
	return that
}

// GetParamArgsFromArgs 从args中获取ParamArgs，同时删除args中的ParamArgs
func GetParamArgsFromArgs(args *[]any) ParamArgs {
	paramArgs := ParamArgs{}
	newArgs := make([]any, 0, len(*args))
	for _, arg := range *args {
		var p ParamArgs
		var ok bool
		if p, ok = arg.(ParamArgs); !ok {
			if p1, ok := arg.(*ParamArgs); ok {
				p = *p1
			}
		}
		if p == nil {
			newArgs = append(newArgs, arg)
			continue
		}
		paramArgs.MergeParamArgs(p)
	}
	*args = newArgs
	return paramArgs
}
