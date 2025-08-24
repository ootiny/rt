package _rt_package_name_

// import (
// 	"database/sql"
// 	"encoding/base64"
// 	"encoding/json"
// 	"regexp"
// 	"strings"

// 	"github.com/google/uuid"
// )

// var (
// 	gSqlCharToNumber = [256]byte{}
// )

// func init() {
// 	for i := byte('0'); i <= '9'; i++ {
// 		gSqlCharToNumber[i] = i - '0'
// 	}

// 	for i := byte('A'); i <= 'F'; i++ {
// 		gSqlCharToNumber[i] = i - 'A' + 10
// 	}

// 	for i := byte('a'); i <= 'f'; i++ {
// 		gSqlCharToNumber[i] = i - 'a' + 10
// 	}
// }

// type SqlQueryOperator string

// const (
// 	SqlEqual        SqlQueryOperator = "="
// 	SqlNotEqual     SqlQueryOperator = "!="
// 	SqlGreaterThan  SqlQueryOperator = ">"
// 	SqlLessThan     SqlQueryOperator = "<"
// 	SqlGreaterEqual SqlQueryOperator = ">="
// 	SqlLessEqual    SqlQueryOperator = "<="
// 	SqlLike         SqlQueryOperator = "like"
// 	SqlIn           SqlQueryOperator = "in"
// 	SqlNotIn        SqlQueryOperator = "not in"
// 	SqlChild        SqlQueryOperator = "child"
// )

// const (
// 	SqlLevelReadCommitted  = "ReadCommitted"
// 	SqlLevelRepeatableRead = "RepeatableRead"
// 	SqlLevelSerializable   = "Serializable"
// )

// func stringToIsolationLevel(s string) sql.IsolationLevel {
// 	switch s {
// 	case SqlLevelReadCommitted:
// 		return sql.LevelReadCommitted
// 	case SqlLevelRepeatableRead:
// 		return sql.LevelRepeatableRead
// 	case SqlLevelSerializable:
// 		return sql.LevelSerializable
// 	default:
// 		return sql.LevelReadCommitted
// 	}
// }

// const gSqlMetaTableName = "_meta_"

// var gSqlIsolationLevels = []string{"", SqlLevelReadCommitted, SqlLevelRepeatableRead, SqlLevelSerializable}
// var gSqlServiceNameRegex, _ = regexp.Compile("^[_a-z][_a-z0-9]*$")
// var gSqlColumnNameRegex, _ = regexp.Compile("^[_a-z][_a-z0-9]*$")
// var gSqlActionNameRegex, _ = regexp.Compile(`^[_a-zA-Z0-9-]+$`)

// var gSqlColumnKindAllowedQueryOperatorsMap = map[string]map[string]bool{
// 	"PK": {
// 		string(SqlEqual):    true,
// 		string(SqlNotEqual): true,
// 		string(SqlIn):       true,
// 		string(SqlNotIn):    true,
// 	},
// 	"LK": {
// 		string(SqlEqual):    true,
// 		string(SqlNotEqual): true,
// 		string(SqlIn):       true,
// 		string(SqlNotIn):    true,
// 	},
// 	"Bool": {
// 		string(SqlEqual):    true,
// 		string(SqlNotEqual): true,
// 		string(SqlIn):       true,
// 		string(SqlNotIn):    true,
// 	},
// 	"Int64": {
// 		string(SqlEqual):        true,
// 		string(SqlNotEqual):     true,
// 		string(SqlGreaterThan):  true,
// 		string(SqlLessThan):     true,
// 		string(SqlGreaterEqual): true,
// 		string(SqlLessEqual):    true,
// 		string(SqlIn):           true,
// 		string(SqlNotIn):        true,
// 	},
// 	"Float64": {
// 		string(SqlGreaterThan):  true,
// 		string(SqlLessThan):     true,
// 		string(SqlGreaterEqual): true,
// 		string(SqlLessEqual):    true,
// 	},
// 	"String": {
// 		string(SqlEqual):    true,
// 		string(SqlNotEqual): true,
// 		string(SqlLike):     true,
// 		string(SqlIn):       true,
// 		string(SqlNotIn):    true,
// 	},
// 	"String256": {
// 		string(SqlEqual):    true,
// 		string(SqlNotEqual): true,
// 		string(SqlLike):     true,
// 		string(SqlIn):       true,
// 		string(SqlNotIn):    true,
// 	},
// 	"Bytes":       {},
// 	"StringArray": {},
// 	"StringMap":   {},
// 	"File":        {},
// 	"FileArray":   {},
// 	"FileMap":     {},
// 	"LKArray":     {},
// 	"LKMap":       {},
// }

// type ISqlAgent interface {
// 	DataSource(host string, port uint16, user string, password string) string
// 	DataSourceWithDB(host string, port uint16, user string, password string, dbName string) string
// 	HasDatabase(dbName string) string
// 	CreateDatabase(dbName string) string
// 	DropDatabase(dbName string) string
// 	CreateGlobalSeed() string
// 	GetGlobalSeed() string
// 	CreateMetaTable() string
// 	QueryMetaTable() string
// 	InsertMetaTable(serviceName string, meta string) string
// 	UpdateMetaTable(serviceName string, meta string) string

// 	CreateServiceTable(serviceName string) string
// 	AddColumn(serviceName string, columnName string, columnType string) string
// 	DropColumn(serviceName string, columnName string) string

// 	CreateIndex(serviceName string, columnName string) string
// 	DropIndex(serviceName string, columnName string) string
// 	CreateUnique(serviceName string, columnName string) string
// 	DropUnique(serviceName string, columnName string) string

// 	Insert(serviceName string, keys []string) string
// 	Update(serviceName string, keys []string) string
// 	Delete(serviceName string) string

// 	QueryOrderBy(serviceName string, query *SqlQuery) string
// 	QueryWhere(serviceName string, argStartPos int, query *SqlQuery) (string, []any, error)
// 	QuerySelect(serviceName string, columns []string) string
// }

// // type ISqlCommands interface {
// // 	Exec(query string, args ...any) (sql.Result, error)
// // 	Query(query string, args ...any) (*sql.Rows, error)
// // 	QueryRow(query string, args ...any) *sql.Row
// // }

// type SqlWhere struct {
// 	columnName string
// 	op         SqlQueryOperator
// 	value      any
// 	concat     string
// }

// func (p *SqlWhere) GetColumnName() string {
// 	return p.columnName
// }

// func (p *SqlWhere) GetOp() SqlQueryOperator {
// 	return p.op
// }

// func (p *SqlWhere) GetValue() any {
// 	return p.value
// }

// func (p *SqlWhere) GetConcat() string {
// 	return p.concat
// }

// type SqlOrderBy struct {
// 	name string
// 	asc  bool
// }

// type SqlQuery struct {
// 	service string
// 	view    string
// 	wheres  []*SqlWhere
// 	orders  []*SqlOrderBy
// 	limit   int
// 	offset  int
// }

// func NewQuery(service string) *SqlQuery {
// 	return &SqlQuery{
// 		service: service,
// 		view:    "",
// 		wheres:  []*SqlWhere{},
// 		orders:  []*SqlOrderBy{},
// 	}
// }

// func (p *SqlQuery) GetService() string {
// 	return p.service
// }

// func (p *SqlQuery) GetView() string {
// 	return p.view
// }

// func (p *SqlQuery) GetWheres() []*SqlWhere {
// 	return p.wheres
// }

// func (p *SqlQuery) GetOrders() []*SqlOrderBy {
// 	return p.orders
// }

// func (p *SqlQuery) GetLimit() int {
// 	return p.limit
// }

// func (p *SqlQuery) GetOffset() int {
// 	return p.offset
// }

// func (p *SqlQuery) View(view string) *SqlQuery {
// 	p.view = view
// 	return p
// }

// func (p *SqlQuery) And(name string, op SqlQueryOperator, value any) *SqlQuery {
// 	p.wheres = append(p.wheres, &SqlWhere{name, op, value, "and"})
// 	return p
// }

// func (p *SqlQuery) Or(name string, op SqlQueryOperator, value any) *SqlQuery {
// 	p.wheres = append(p.wheres, &SqlWhere{name, op, value, "or"})
// 	return p
// }

// func (p *SqlQuery) AndChild(child *SqlQuery) *SqlQuery {
// 	p.wheres = append(p.wheres, &SqlWhere{"", SqlChild, child, "and"})
// 	return p
// }

// func (p *SqlQuery) OrChild(child *SqlQuery) *SqlQuery {
// 	p.wheres = append(p.wheres, &SqlWhere{"", SqlChild, child, "or"})
// 	return p
// }

// func (p *SqlQuery) OrderByAsc(name string) *SqlQuery {
// 	p.orders = append(p.orders, &SqlOrderBy{name, true})
// 	return p
// }

// func (p *SqlQuery) OrderByDesc(name string) *SqlQuery {
// 	p.orders = append(p.orders, &SqlOrderBy{name, false})
// 	return p
// }

// func (p *SqlQuery) Limit(limit int) *SqlQuery {
// 	p.limit = limit
// 	return p
// }

// func (p *SqlQuery) Offset(offset int) *SqlQuery {
// 	p.offset = offset
// 	return p
// }

// func (p *SqlQuery) Check(service *SqlServiceMeta, root bool) error {
// 	// check orderBy
// 	for _, v := range p.orders {
// 		if !service.CanOrder(v.name) {
// 			return Errorf("Query.Check: %s.%s order is not allowed", service.GetName(), v.name)
// 		}
// 	}

// 	// check where
// 	for _, v := range p.wheres {
// 		if v.op == SqlChild {
// 			if e := v.value.(*SqlQuery).Check(service, false); e != nil {
// 				return WrapError(e)
// 			}
// 		} else {
// 			if !service.CanQuery(v.columnName, v.op) {
// 				return Errorf("Query.Check: %s.%s \"%s\"is not allowed", service.GetName(), v.columnName, v.op)
// 			}
// 		}
// 	}

// 	// check limit and offset
// 	if p.limit < 0 || p.offset < 0 {
// 		return Errorf("Query.Check: limit and offset must be positive")
// 	}

// 	return nil
// }

// func NewWebQuery(name string, queries map[string]any, orders []string) *SqlQuery {
// 	ret := NewQuery(name)
// 	for key, value := range queries {
// 		// key := "level:in"
// 		arr := strings.Split(key, ":")
// 		if len(arr) != 2 {
// 			continue
// 		}
// 		ret.And(arr[0], SqlQueryOperator(arr[1]), value)
// 	}

// 	for _, order := range orders {
// 		// order := "level:ascend"
// 		arr := strings.Split(order, ":")
// 		if len(arr) != 2 {
// 			continue
// 		}
// 		if arr[1] == "ascend" {
// 			ret.OrderByAsc(arr[0])
// 		} else if arr[1] == "descend" {
// 			ret.OrderByDesc(arr[0])
// 		} else {
// 			continue
// 		}
// 	}

// 	return ret
// }

// type SqlColumnConfig struct {
// 	Type     string   `json:"type"`
// 	Query    []string `json:"query"`
// 	Unique   bool     `json:"unique"`
// 	Index    bool     `json:"index"`
// 	Order    bool     `json:"order"`
// 	mapQuery map[string]bool
// }

// type SqlActionConfig struct {
// 	UseSyncLock bool   `json:"useSyncLock"`
// 	Editable    bool   `json:"editable"`
// 	Tx          string `json:"tx"`
// }

// type linkConfig struct {
// 	service string
// 	view    string
// }

// type SqlViewConfig struct {
// 	Cache        string   `json:"cache"`
// 	Columns      []string `json:"columns"`
// 	mapLink      map[string]*linkConfig
// 	arraySelects []string
// 	cacheSecond  int64
// }

// func (p *SqlViewConfig) GetCacheSecond() int64 {
// 	return p.cacheSecond
// }

// func SqlEncodeToDB(kind string, v any) (any, error) {
// 	switch kind {
// 	case "PK", "LK", "Bool", "Int64", "Float64", "String", "String256", "File", "Bytes":
// 		return v, nil
// 	case "StringArray", "LKArray", "FileArray":
// 		if v == nil {
// 			return "[]", nil
// 		}
// 	case "StringMap", "LKMap", "FileMap":
// 		if v == nil {
// 			return "{}", nil
// 		}
// 	default:
// 		return v, Errorf("unknown kind %s", kind)
// 	}

// 	if ret, e := json.Marshal(v); e != nil {
// 		return nil, WrapError(e)
// 	} else {
// 		return string(ret), nil
// 	}
// }

// func SqlDecodeFromDB(kind string, v any) (any, error) {
// 	switch kind {
// 	case "PK", "LK", "Bool", "Int64", "Float64", "String", "String256", "File", "Bytes":
// 		return v, nil
// 	case "StringArray", "LKArray", "FileArray":
// 		ret := make([]string, 0)
// 		if v == nil {
// 			return make([]string, 0), nil
// 		} else if strV, ok := v.(string); !ok {
// 			return ret, WrapError(ErrSystem)
// 		} else if e := json.Unmarshal([]byte(strV), &ret); e != nil {
// 			return ret, WrapError(e)
// 		} else {
// 			return ret, nil
// 		}
// 	case "StringMap":
// 		ret := make(map[string]any)
// 		if v == nil {
// 			return make(map[string]any), nil
// 		} else if strV, ok := v.(string); !ok {
// 			return ret, WrapError(ErrSystem)
// 		} else if e := json.Unmarshal([]byte(strV), &ret); e != nil {
// 			return ret, WrapError(e)
// 		} else {
// 			return ret, nil
// 		}
// 	case "LKMap", "FileMap":
// 		ret := make(map[string]string)
// 		if v == nil {
// 			return ret, nil
// 		} else if strV, ok := v.(string); !ok {
// 			return ret, WrapError(ErrSystem)
// 		} else if e := json.Unmarshal([]byte(strV), &ret); e != nil {
// 			return ret, WrapError(e)
// 		} else {
// 			return ret, nil
// 		}
// 	default:
// 		return nil, Errorf("unknown kind %s", kind)
// 	}
// }

// func SqlRowsToRecords(service *SqlServiceMeta, viewSelects []string, rows *sql.Rows) ([]Record, error) {
// 	ret := make([]Record, 0)

// 	if len(viewSelects) == 0 {
// 		return nil, Errorf("viewSelects is empty")
// 	}

// 	for rows.Next() {
// 		item := Record{}
// 		scanArgs := make([]any, len(viewSelects))
// 		columnConfig := service.GetColumnsMap()
// 		for i := 0; i < len(scanArgs); i++ {
// 			var v any
// 			scanArgs[i] = &v
// 		}

// 		if e := rows.Scan(scanArgs...); e != nil {
// 			return nil, WrapError(e, "rows.Scan error")
// 		}

// 		for i, column := range viewSelects {
// 			if v, e := SqlDecodeFromDB(columnConfig[column], *(scanArgs[i].(*any))); e != nil {
// 				return nil, WrapError(e, "SqlDecodeFromDB error")
// 			} else {
// 				item[column] = v
// 			}
// 		}

// 		ret = append(ret, item)
// 	}

// 	if e := rows.Err(); e != nil { // Check for any error occurred during iteration
// 		return nil, WrapError(e, "rows iteration error")
// 	}

// 	return ret, nil
// }

// func SqlUUID() string {
// 	raw := []byte(strings.Replace(uuid.NewString(), "-", "", -1))
// 	buffer := make([]byte, 16)
// 	for i := 0; i < 32; i += 2 {
// 		buffer[i/2] = gSqlCharToNumber[raw[i]]*16 + gSqlCharToNumber[raw[i+1]]
// 	}
// 	return base64.RawURLEncoding.EncodeToString(buffer)
// }

// // SqlDiffStringArray returns the elements to add and to delete to convert left to right.
// func SqlDiffStringArray(left []string, right []string) ([]string, []string) {
// 	toDelete := []string{}
// 	toAdd := []string{}

// 	leftSet := make(map[string]struct{})
// 	rightSet := make(map[string]struct{})

// 	for _, item := range left {
// 		leftSet[item] = struct{}{}
// 	}
// 	for _, item := range right {
// 		rightSet[item] = struct{}{}
// 	}

// 	for item := range leftSet {
// 		if _, found := rightSet[item]; !found {
// 			toDelete = append(toDelete, item)
// 		}
// 	}

// 	for item := range rightSet {
// 		if _, found := leftSet[item]; !found {
// 			toAdd = append(toAdd, item)
// 		}
// 	}

// 	return toAdd, toDelete
// }

// // SqlDiffStringMap returns the elements to add, to delete, and to change to convert left to right.
// func SqlDiffStringMap(
// 	left map[string]string,
// 	right map[string]string,
// ) (map[string]string, map[string]string, map[string]string) {
// 	toAdd := make(map[string]string)
// 	toDelete := make(map[string]string)
// 	toChange := make(map[string]string)

// 	for key, leftValue := range left {
// 		if rightValue, found := right[key]; !found {
// 			toDelete[key] = leftValue
// 		} else if leftValue != rightValue {
// 			toChange[key] = rightValue
// 		}
// 	}

// 	for key, rightValue := range right {
// 		if _, found := left[key]; !found {
// 			toAdd[key] = rightValue
// 		}
// 	}

// 	return toAdd, toChange, toDelete
// }

// func ArrayContains(arr []string, value string) bool {
// 	for _, v := range arr {
// 		if v == value {
// 			return true
// 		}
// 	}
// 	return false
// }

// func UniqueArrayWithoutEmpty(arr []string) []string {
// 	ret := make([]string, 0)
// 	set := make(map[string]struct{})

// 	for _, v := range arr {
// 		if v != "" {
// 			if _, ok := set[v]; !ok {
// 				set[v] = struct{}{}
// 				ret = append(ret, v)
// 			}
// 		}
// 	}

// 	return ret
// }
