package _rt_package_name_

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
)

type SQLTransaction struct {
	readOnly       bool
	tx             *sql.Tx
	dbMgr          *SQLManager
	isolationLevel string
	mutex          *sync.Mutex
}

func (p *SQLTransaction) GetTx() (*sql.Tx, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.tx == nil {
		if tx, e := p.dbMgr.db.BeginTx(context.Background(), &sql.TxOptions{
			Isolation: stringToIsolationLevel(p.isolationLevel),
			ReadOnly:  p.readOnly,
		}); e != nil {
			return nil, e
		} else {
			p.tx = tx
		}
	}

	return p.tx, nil
}

func (p *SQLTransaction) Close(commit bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	ret := error(nil)

	if p.tx != nil {
		if commit {
			ret = p.tx.Commit()
		} else {
			ret = p.tx.Rollback()
		}
		p.tx = nil
	}

	return ret
}

func (p *SQLTransaction) Add(serviceName string, record Record) error {
	agent := p.dbMgr.agent
	service := p.dbMgr.GetService(serviceName)

	if service == nil {
		return Errorf("Add: service %s not found", serviceName)
	}

	for columnName, columnKind := range service.GetColumnsMap() {
		if _, ok := record[columnName]; !ok {
			switch columnKind {
			case "PK":
				record[columnName] = SqlUUID()
			case "LK":
				record[columnName] = ""
			case "Bool":
				record[columnName] = false
			case "Int64":
				record[columnName] = int64(0)
			case "Float64":
				record[columnName] = float64(0)
			case "Bytes":
				record[columnName] = []byte{}
			case "String256":
				record[columnName] = ""
			case "String":
				record[columnName] = ""
			case "StringArray":
				record[columnName] = []any{}
			case "StringMap":
				record[columnName] = map[string]any{}
			case "File":
				record[columnName] = ""
			case "FileArray":
				record[columnName] = []string{}
			case "FileMap":
				record[columnName] = map[string]string{}
			case "LKArray":
				record[columnName] = []string{}
			case "LKMap":
				record[columnName] = map[string]string{}
			default:
			}
		}
	}

	pos := 0
	keys := make([]string, 0)
	arguments := make([]any, 0)

	columnConfig := service.GetColumnsMap()

	for columnName, columnValue := range record {
		pos++
		keys = append(keys, columnName)

		kind := columnConfig[columnName]

		if kind == "" {
			return Errorf("Add: %s: unknown column \"%s\"", service.GetName(), columnName)
		} else if argValue, e := SqlEncodeToDB(kind, columnValue); e != nil {
			return WrapError(e, "Add", service.GetName())
		} else {
			arguments = append(arguments, argValue)
		}
	}

	if len(keys) == 0 {
		return nil
	}

	if tx, e := p.GetTx(); e != nil {
		return WrapError(e, "Add", service.GetName())
	} else if _, e := tx.Exec(
		agent.Insert(service.GetName(), keys),
		arguments...,
	); e != nil {
		return WrapError(e, "Add", service.GetName())
	} else {
		return nil
	}
}

func (p *SQLTransaction) Del(serviceName string, id string) error {
	agent := p.dbMgr.agent
	service := p.dbMgr.GetService(serviceName)

	if service == nil {
		return Errorf("Delete: %s: service is ", serviceName)
	}

	if tx, e := p.GetTx(); e != nil {
		return WrapError(e, "Delete", service.GetName())
	} else if _, e := tx.Exec(
		agent.Delete(service.GetName()),
		id,
	); e != nil {
		return WrapError(e, "Delete", service.GetName())
	} else {
		return nil
	}
}

func (p *SQLTransaction) Update(serviceName string, record Record) error {
	agent := p.dbMgr.agent
	service := p.dbMgr.GetService(serviceName)

	if service == nil {
		return Errorf("Update: service %s not found", serviceName)
	}

	if recordId, ok := record.String("id"); !ok {
		return Errorf("Update: %s: record has no id", serviceName)
	} else {
		columnConfig := service.GetColumnsMap()
		keys := make([]string, 0)
		arguments := make([]any, 0)

		for columnName, columnValue := range record {
			if columnName != "id" {
				if len(columnConfig[columnName]) <= 0 {
					return Errorf("Update: %s: invalid column %s", service.GetName(), columnName)
				}

				if v, e := SqlEncodeToDB(columnConfig[columnName], columnValue); e != nil {
					return WrapError(e, "Update", service.GetName())
				} else {
					keys = append(keys, columnName)
					arguments = append(arguments, v)
				}
			}
		}

		if len(keys) == 0 {
			return nil
		}

		arguments = append(arguments, recordId)

		if tx, e := p.GetTx(); e != nil {
			return WrapError(e, "Update", service.GetName())
		} else if _, e := tx.Exec(
			agent.Update(service.GetName(), keys),
			arguments...,
		); e != nil {
			return WrapError(e, "Update", service.GetName())
		} else {
			return nil
		}
	}
}

func (p *SQLTransaction) GetColumnType(serviceName string, columnName string) string {
	service := p.dbMgr.GetService(serviceName)

	if service == nil {
		return ""
	}

	return service.GetColumnType(columnName)
}

func (p *SQLTransaction) GetRecordWithAllColumns(serviceName string, id string) (ret Record, err error) {
	agent := p.dbMgr.agent
	service := p.dbMgr.GetService(serviceName)

	if service == nil {
		return Record{}, Errorf("GetColumnValue: service %s not found", serviceName)
	}

	viewSelects := []string{}
	for columnName := range service.GetColumnsMap() {
		viewSelects = append(viewSelects, columnName)
	}

	query := NewQuery(serviceName).And("id", "=", id)

	// check and build where
	execWhere, whereArgs, e := agent.QueryWhere(service.GetName(), 0, query)
	if e != nil {
		return Record{}, WrapError(e, "GetColumnValue", service.GetName())
	}

	if execWhere != "" {
		execWhere = "WHERE " + execWhere
	}

	if tx, e := p.GetTx(); e != nil {
		return Record{}, WrapError(e, "GetColumnValue", service.GetName())
	} else if rows, e := tx.Query(fmt.Sprintf(
		"SELECT %s FROM \"%s\" %s;",
		agent.QuerySelect(service.GetName(), viewSelects),
		service.GetName(),
		execWhere,
	), whereArgs...); e != nil {
		return Record{}, WrapError(e, "GetColumnValue", service.GetName())
	} else {
		defer func() {
			if rows != nil {
				err = GetFirstError(err, WrapError(rows.Close()))
			}
		}()

		if ret, e := SqlRowsToRecords(service, viewSelects, rows); e != nil {
			return Record{}, Errorf("service %s GetColumnValue error: %s", service.GetName(), e.Error())
		} else if len(ret) != 1 {
			return Record{}, Errorf("service %s GetColumnValue error: record not found", service.GetName())
		} else {
			return ret[0], nil
		}
	}
}

func (p *SQLTransaction) Query(query *SqlQuery) (ret []Record, err error) {
	serviceName := query.GetService()
	agent := p.dbMgr.agent
	service := p.dbMgr.GetService(serviceName)

	if service == nil {
		return []Record{}, Errorf("Query: service %s not found", serviceName)
	}

	if query == nil {
		return []Record{}, Errorf("Query: %s: service is nil", serviceName)
	}

	if e := query.Check(service, true); e != nil {
		return []Record{}, WrapError(e, "Query", service.GetName())
	}

	viewSelects := service.GetViewSelects(query.GetView())
	if len(viewSelects) == 0 {
		return []Record{}, Errorf("Query: %s: view %s not found", service.GetName(), query.GetView())
	}

	// check and build where
	execWhere, whereArgs, e := agent.QueryWhere(service.GetName(), 0, query)
	if e != nil {
		return []Record{}, WrapError(e, "Query", service.GetName())
	}

	execOrderBy := agent.QueryOrderBy(service.GetName(), query)
	execLimit := ""

	if query.GetLimit() > 0 && query.GetOffset() >= 0 {
		execLimit = fmt.Sprintf("LIMIT %d OFFSET %d", query.GetLimit(), query.GetOffset())
	} else if query.GetLimit() > 0 {
		execLimit = fmt.Sprintf("LIMIT %d", query.GetLimit())
	} else if query.GetOffset() >= 0 {
		execLimit = fmt.Sprintf("OFFSET %d", query.GetOffset())
	} else {
		execLimit = ""
	}

	if execWhere != "" {
		execWhere = "WHERE " + execWhere
	}

	if execOrderBy != "" {
		execOrderBy = "ORDER BY " + execOrderBy
	}

	// fmt.Printf(
	// 	"SELECT %s FROM \"%s\" %s %s %s;\n",
	// 	agent.QuerySelect(service.GetName(), viewSelects),
	// 	service.GetName(),
	// 	execWhere,
	// 	execOrderBy,
	// 	execLimit,
	// )

	if tx, e := p.GetTx(); e != nil {
		return []Record{}, WrapError(e, "Query", service.GetName())
	} else if rows, e := tx.Query(fmt.Sprintf(
		"SELECT %s FROM \"%s\" %s %s %s;",
		agent.QuerySelect(service.GetName(), viewSelects),
		service.GetName(),
		execWhere,
		execOrderBy,
		execLimit,
	), whereArgs...); e != nil {
		return []Record{}, WrapError(e, "Query", service.GetName())
	} else {
		defer func() {
			if rows != nil {
				err = GetFirstError(err, WrapError(rows.Close()))
			}
		}()

		if ret, e := SqlRowsToRecords(service, viewSelects, rows); e != nil {
			return []Record{}, Errorf("service %s Query error: %s", service.GetName(), e.Error())
		} else {
			return ret, nil
		}
	}
}

func (p *SQLTransaction) GlobalSeed() (int64, error) {
	val := int64(0)

	if tx, e := p.GetTx(); e != nil {
		return -1, Errorf("GlobalSeed error: %s", e.Error())
	} else if e := tx.QueryRow(p.dbMgr.agent.GetGlobalSeed()).Scan(&val); e != nil {
		return -1, Errorf("GlobalSeed error: %s", e.Error())
	} else {
		return val, nil
	}
}

func (p *SQLTransaction) UpdateTable(newConfigText string) error {
	agent := p.dbMgr.agent
	service, e := NewSqlServiceMeta(newConfigText)

	if e != nil {
		return WrapError(e)
	}

	dbServiceMetaStr := ""

	tx, e := p.GetTx()
	if e != nil {
		return WrapError(e)
	}

	if _, e := tx.Exec(
		agent.CreateMetaTable(),
	); e != nil {
		return WrapError(e)
	}

	if _, e := tx.Exec(agent.CreateGlobalSeed()); e != nil {
		return WrapError(e)
	}

	if rows, e := tx.Query(agent.QueryMetaTable(), service.GetName()); e != nil {
		return WrapError(e)
	} else {
		for rows.Next() {
			if e := rows.Scan(&dbServiceMetaStr); e != nil {
				return WrapError(e)
			}
		}
	}

	dbServiceMeta := (*SqlServiceMeta)(nil)
	if dbServiceMetaStr != "" {
		dbServiceMeta, e = NewSqlServiceMeta(dbServiceMetaStr)
		if e != nil {
			return WrapError(e, fmt.Sprintf("database service meta %s", service.GetName()))
		}
	} else {
		dbServiceMeta = &SqlServiceMeta{}
	}

	addColumns, changeColumns, delColumns := SqlDiffStringMap(dbServiceMeta.GetColumnsMap(), service.GetColumnsMap())

	if len(changeColumns) != 0 {
		return Errorf(
			"UpdateTable %s: Columns: new meta is not incompatible with base meta",
			service.GetName(),
		)
	}

	if len(delColumns) != 0 {
		for columnName := range delColumns {
			fmt.Printf(
				"Table \"%s\" column \"%s\" will be deleted, please input name to confirm: ",
				service.GetName(), columnName,
			)

			if inputText, e := bufio.NewReader(os.Stdin).ReadString('\n'); e != nil {
				return WrapError(e)
			} else if strings.TrimSpace(inputText) != columnName {
				return Errorf("UpdateTable %s: canceled by user", service.GetName())
			} else {
				fmt.Printf(
					"Confirm to delete column \"%s\" from table \"%s\" ? ",
					columnName, service.GetName(),
				)
				if inputText, e := bufio.NewReader(os.Stdin).ReadString('\n'); e != nil {
					return WrapError(e)
				} else if strings.TrimSpace(inputText) != "Y" {
					return Errorf("UpdateTable %s: canceled by user", service.GetName())
				}

				// delete columns
				if _, e := tx.Exec(
					agent.DropColumn(service.GetName(), columnName),
				); e != nil {
					return WrapError(e)
				}
			}
		}
	}

	// prepare init sql
	execList := []string{}
	if len(service.Columns) > 0 {
		execList = append(execList, agent.CreateServiceTable(service.GetName()))
	}

	// add columns
	for columnName, columnKind := range addColumns {
		switch columnKind {
		case "PK":
			break
		default:
			sqlStr := agent.AddColumn(service.GetName(), columnName, columnKind)
			if sqlStr == "" {
				return Errorf("invalid column kind %s", columnKind)
			} else {
				execList = append(execList, sqlStr)
			}
		}
	}

	// update indexes
	addIndexes, delIndexes := SqlDiffStringArray(dbServiceMeta.GetIndexes(), service.GetIndexes())

	for _, columnName := range delIndexes {
		execList = append(execList, agent.DropIndex(service.GetName(), columnName))
	}
	for _, columnName := range addIndexes {
		execList = append(execList, agent.CreateIndex(service.GetName(), columnName))
	}

	// update uniques
	addUniques, delUniques := SqlDiffStringArray(dbServiceMeta.GetUniques(), service.GetUniques())

	for _, columnName := range delUniques {
		execList = append(execList, agent.DropUnique(service.GetName(), columnName))
	}
	for _, columnName := range addUniques {
		execList = append(execList, agent.CreateUnique(service.GetName(), columnName))
	}

	// update meta
	if dbServiceMetaStr == "" {
		execList = append(execList, agent.InsertMetaTable(service.GetName(), newConfigText))
	} else {
		execList = append(execList, agent.UpdateMetaTable(service.GetName(), newConfigText))
	}

	// exec sql
	for i := 0; i < len(execList); i++ {
		if _, e := tx.Exec(execList[i]); e != nil {
			return WrapError(e)
		}
	}

	return nil
}
