package builder

import (
	"fmt"
	"strings"
)

func ParseColumnConfig(columnConfig map[string]interface{}) *DBColumnConfig {
	column := &DBColumnConfig{}
	column.Type = columnConfig["type"].(string)
	column.Query = columnConfig["query"].([]string)
	column.Unique = columnConfig["unique"].(bool)
	column.Index = columnConfig["index"].(bool)
	column.Order = columnConfig["order"].(bool)
	column.LinkTable = columnConfig["linkTable"].(string)
	return column
}

type DBColumnConfig struct {
	Type      string   `json:"type"`
	Query     []string `json:"query"`
	Unique    bool     `json:"unique"`
	Index     bool     `json:"index"`
	Order     bool     `json:"order"`
	LinkTable string   `json:"linkTable"`
	Required  bool     `json:"required"`
	mapQuery  map[string]bool
}

func (p *DBColumnConfig) CanQuery(column string) bool {
	return p.mapQuery[column]
}

type DBViewColumnConfig struct {
	Name      string `json:"name"`
	LinkTable string `json:"linkTable"`
	LinkView  string `json:"linkView"`
}

type DBViewConfig struct {
	Cache       string   `json:"cache"`
	Columns     []string `json:"columns"`
	querySelect string
	cacheSecond int64
}

func (p *DBViewConfig) GetQuerySelect() string {
	return p.querySelect
}

func (p *DBViewConfig) GetCacheSecond() int64 {
	return p.cacheSecond
}

type DBTableConfig struct {
	Version      string                     `json:"version"`
	Table        string                     `json:"table"`
	Columns      map[string]*DBColumnConfig `json:"columns"`
	Views        map[string]*DBViewConfig   `json:"views"`
	__filepath__ string
}

func (p *DBTableConfig) GetFilePath() string {
	return p.__filepath__
}

func (p *DBTableConfig) ToApiConfig() (*APIConfig, error) {
	fnDBTypeToAPIType := func(v string, klass string) string {
		switch v {
		case "PK":
			return "String"
		case "Bool":
			return "Bool"
		case "Int64":
			return "Int64"
		case "Float64":
			return "Float64"
		case "String", "String32", "String64", "String256":
			return "String"
		case "List<String>":
			return "List<String>"
		case "Map<String>":
			return "Map<String>"
		case "Bytes":
			return "Bytes"
		default:
			if strings.HasPrefix(v, "List<") && strings.HasSuffix(v, ">") {
				innerType := v[5 : len(v)-1]
				return fmt.Sprintf("List<%s@%s>", innerType, klass)
			} else if strings.HasPrefix(v, "Map<") && strings.HasSuffix(v, ">") {
				innerType := v[4 : len(v)-1]
				return fmt.Sprintf("Map<%s@%s>", innerType, klass)
			} else if strings.HasPrefix(v, DBPrefix) {
				return fmt.Sprintf("%s@%s", v, klass)
			} else {
				return v
			}
		}
	}

	definitions := map[string]*APIDefinitionConfig{}

	for name, view := range p.Views {
		attributes := []*APIDefinitionAttributeConfig{}

		for _, column := range view.Columns {
			columnName := ""
			columnType := ""
			columnArray := strings.Split(column, "@")
			if len(columnArray) == 1 {
				columnName = columnArray[0]
				columnType = fnDBTypeToAPIType(p.Columns[columnName].Type, "")
			} else {
				columnName = columnArray[0]
				columnType = fnDBTypeToAPIType(p.Columns[columnName].Type, columnArray[1])
			}

			attributes = append(attributes, &APIDefinitionAttributeConfig{
				Name:     columnName,
				Type:     columnType,
				Required: p.Columns[columnName].Required,
			})
		}
		definitions[name] = &APIDefinitionConfig{
			Attributes: attributes,
		}
	}

	return &APIConfig{
		Version:     "rt.db.v1",
		Namespace:   p.Table,
		Definitions: definitions,
	}, nil
}

func LoadDBTableConfig(filePath string) (*DBTableConfig, error) {
	var config DBTableConfig
	if err := UnmarshalConfig(filePath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	} else {
		config.__filepath__ = filePath
		return &config, nil
	}
}
