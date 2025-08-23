package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type APIDefinitionAttributeConfig struct {
	Name        string `json:"name" required:"true"`
	Type        string `json:"type" required:"true"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIDefinitionConfig struct {
	Description string                          `json:"description"`
	Attributes  []*APIDefinitionAttributeConfig `json:"attributes"`
}

type APIActionParameterConfig struct {
	Name        string `json:"name" required:"true"`
	Type        string `json:"type" required:"true"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIActionReturnConfig struct {
	Type        string `json:"type" required:"true"`
	Description string `json:"description"`
}

type APIActionConfig struct {
	Description string                      `json:"description"`
	Method      string                      `json:"method" required:"true"`
	Parameters  []*APIActionParameterConfig `json:"parameters"`
	Return      *APIActionReturnConfig      `json:"return"`
}

type APIConfig struct {
	Version      string                          `json:"version" required:"true"`
	Namespace    string                          `json:"namespace" required:"true"`
	Description  string                          `json:"description"`
	Definitions  map[string]*APIDefinitionConfig `json:"definitions" required:"true"`
	Actions      map[string]*APIActionConfig     `json:"actions"`
	__filepath__ string
}

func (c *APIConfig) GetFilePath() string {
	return c.__filepath__
}

type APIConfigNode struct {
	name      string
	namespace string
	config    *APIConfig
	children  map[string]*APIConfigNode
}

func MakeApiConfigTree(configlist []*APIConfig) *APIConfigNode {
	nsMap := map[string]*APIConfig{}
	for _, config := range configlist {
		nsMap[config.Namespace] = config
	}

	buildMap := map[string]*APIConfigNode{}

	for _, config := range nsMap {
		nsArr := strings.Split(config.Namespace, ".")

		if len(nsArr) > 1 && (nsArr[0] == "API" || nsArr[0] == "DB") {
			for i := range nsArr {
				partNS := strings.Join(nsArr[:i+1], ".")
				if _, ok := buildMap[partNS]; !ok {
					buildMap[partNS] = &APIConfigNode{
						name:      nsArr[i],
						namespace: partNS,
						config:    nil,
						children:  map[string]*APIConfigNode{},
					}
				}
			}

			buildMap[config.Namespace].config = config
		}
	}

	// 建立父子关系
	for namespace, node := range buildMap {
		nsArr := strings.Split(namespace, ".")
		if len(nsArr) > 1 {
			// 找到父节点的namespace
			parentNS := strings.Join(nsArr[:len(nsArr)-1], ".")
			if parentNode, ok := buildMap[parentNS]; ok {
				// 将当前节点添加到父节点的children中
				parentNode.children[node.name] = node
			}
		}
	}

	return &APIConfigNode{
		name:      "",
		namespace: "",
		config:    nil,
		children: map[string]*APIConfigNode{
			"API": buildMap["API"],
			"DB":  buildMap["DB"],
		},
	}
}

func LoadAPIConfig(filePath string) (*APIConfig, error) {
	var config APIConfig

	if err := UnmarshalConfig(filePath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.__filepath__ = filePath

	return &config, nil
}

type TableColumnConfig struct {
	Type     string   `json:"type"`
	Query    []string `json:"query"`
	Unique   bool     `json:"unique"`
	Index    bool     `json:"index"`
	Order    bool     `json:"order"`
	Required bool     `json:"required"`
}

func (p *TableColumnConfig) ToDBTableColumn() (*DBTableColumn, error) {
	// parse type
	strType := ""
	strTable := ""
	switch p.Type {
	case "PK", "Bool", "Int64", "Float64",
		"String", "String16", "String32", "String64", "String256",
		"List<String>", "Map<String>":
		strType = p.Type
		strTable = ""
	default:
		if strings.HasPrefix(p.Type, DBPrefix) {
			strType = "LK"
			strTable = NamespaceToTableName(p.Type)
		} else if strings.HasPrefix(p.Type, "List<") && strings.HasSuffix(p.Type, ">") {
			innerType := p.Type[5 : len(p.Type)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				strType = "LKList"
				strTable = NamespaceToTableName(innerType)
			} else {
				return nil, fmt.Errorf("invalid column type: %s", p.Type)
			}
		} else if strings.HasPrefix(p.Type, "Map<") && strings.HasSuffix(p.Type, ">") {
			innerType := p.Type[4 : len(p.Type)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				strType = "LKMap"
				strTable = NamespaceToTableName(innerType)
			} else {
				return nil, fmt.Errorf("invalid column type: %s", p.Type)
			}
		} else {
			return nil, fmt.Errorf("invalid column type: %s", p.Type)
		}
	}

	// build query map
	queryMap := map[string]bool{}
	for _, v := range p.Query {
		queryMap[v] = true
	}

	return &DBTableColumn{
		Type:      strType,
		QueryMap:  queryMap,
		Unique:    p.Unique,
		Index:     p.Index,
		Order:     p.Order,
		Required:  p.Required,
		LinkTable: strTable,
	}, nil
}

type TableViewConfig struct {
	Cache   string   `json:"cache"`
	Columns []string `json:"columns"`
}

type TableConfig struct {
	Version      string                        `json:"version"`
	Table        string                        `json:"table"`
	Columns      map[string]*TableColumnConfig `json:"columns"`
	Views        map[string]*TableViewConfig   `json:"views"`
	__filepath__ string
}

func (p *TableConfig) GetFilePath() string {
	return p.__filepath__
}

func (p *TableConfig) ToApiConfig() (*APIConfig, error) {
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

func (p *TableConfig) ToDBTable() (*DBTable, error) {
	// convert columns
	columns := map[string]*DBTableColumn{}
	for name, column := range p.Columns {
		if dbColumn, err := column.ToDBTableColumn(); err != nil {
			return nil, err
		} else {
			columns[name] = dbColumn
		}
	}

	viewNames := []string{}
	for name := range p.Views {
		viewNames = append(viewNames, name)
	}
	sort.Strings(viewNames)

	// convert views
	views := map[string]*DBTableView{}
	for name, view := range p.Views {
		viewHashArr := []string{}
		viewColumns := []*DBTableViewColumn{}

		// get view index by viewNames
		viewIndex := uint64(0)
		for i, viewName := range viewNames {
			if viewName == name {
				viewIndex = uint64(i)
				break
			}
		}

		for _, viewColumn := range view.Columns {
			columnArray := strings.Split(viewColumn, "@")
			if column, ok := columns[columnArray[0]]; !ok {
				return nil, fmt.Errorf("views.%s column %s not found", name, columnArray[0])
			} else {
				if len(columnArray) == 1 {
					viewColumns = append(viewColumns, &DBTableViewColumn{
						Name:      columnArray[0],
						LinkTable: "",
						LinkView:  "",
					})
					viewHashArr = append(viewHashArr, columnArray[0])
				} else if len(columnArray) == 2 {
					viewColumns = append(viewColumns, &DBTableViewColumn{
						Name:      columnArray[0],
						LinkTable: column.LinkTable,
						LinkView:  columnArray[1],
					})
					viewHashArr = append(viewHashArr, columnArray[0]+"@"+column.LinkTable+"@"+columnArray[1])
				} else {
					return nil, fmt.Errorf("views.%s column %s invalid", name, viewColumn)
				}
			}
		}

		// sort viewHashArr
		sort.Strings(viewHashArr)

		// build columnsSelect
		columnsSelectList := []string{}
		for _, viewColumn := range viewColumns {
			columnsSelectList = append(columnsSelectList, viewColumn.Name)
		}

		if len(columnsSelectList) == 0 {
			return nil, fmt.Errorf("views.%s has no columns", name)
		}

		// parse cache
		cacheSecond, err := TimeStringToDuration(view.Cache)
		if err != nil {
			return nil, fmt.Errorf("views.%s cache invalid: %w", name, err)
		}

		// make md5Hash
		views[name] = &DBTableView{
			Columns:       viewColumns,
			ColumnsSelect: strings.Join(columnsSelectList, ","),
			CacheSecond:   int64(cacheSecond / time.Second),
			Hash:          GetViewHash(viewIndex+1, strings.Join(viewHashArr, ":")),
		}
	}

	return &DBTable{
		Version:   "rt.dbservice.v1",
		Table:     NamespaceToTableName(p.Table),
		Columns:   columns,
		Views:     views,
		Namespace: p.Table,
		File:      p.GetFilePath(),
	}, nil
}

func LoadTableConfig(filePath string) (*TableConfig, error) {
	var config TableConfig
	if err := UnmarshalConfig(filePath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	} else {
		config.__filepath__ = filePath
		return &config, nil
	}
}

type RTOutputConfig struct {
	Kind       string `json:"kind" required:"true"`
	Language   string `json:"language" required:"true"`
	Dir        string `json:"dir" required:"true"`
	GoModule   string `json:"goModule"`
	GoPackage  string `json:"goPackage"`
	HttpEngine string `json:"httpEngine"`
}

type RTDatabaseConfig struct {
	Driver    string `json:"driver" required:"true"`
	Host      string `json:"host" required:"true"`
	Port      uint16 `json:"port" required:"true"`
	User      string `json:"user" required:"true"`
	Password  string `json:"password" required:"true"`
	DBName    string `json:"dbName" required:"true"`
	CacheSize string `json:"cacheSize"`
}

type RTConfig struct {
	Listen       string           `json:"listen"`
	Outputs      []RTOutputConfig `json:"outputs"`
	Database     RTDatabaseConfig `json:"database"`
	__filepath__ string
}

func (c *RTConfig) GetFilePath() string {
	return c.__filepath__
}

func LoadRtConfig() (*RTConfig, error) {
	configPath := ""

	if len(os.Args) > 1 {
		if fileInfo, err := os.Stat(os.Args[1]); err == nil && !fileInfo.IsDir() {
			configPath = os.Args[1]
		}
	}

	if configPath == "" {
		// 在当前目录下，依次寻找 .rt.json .rt.yaml .rt.yml
		searchFiles := []string{"./.rt.json", "./.rt.yaml", "./.rt.yml"}
		for _, file := range searchFiles {
			if fileInfo, err := os.Stat(file); err == nil && !fileInfo.IsDir() {
				configPath = file
				break
			}
		}
	}

	if !filepath.IsAbs(configPath) {
		if absPath, err := filepath.Abs(configPath); err != nil {
			return nil, fmt.Errorf("failed to convert config path to absolute path: %v", err)
		} else {
			configPath = absPath
		}
	}

	var config RTConfig

	if err := UnmarshalConfig(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	projectDir := filepath.Dir(configPath)

	for i, output := range config.Outputs {
		var err error
		config.Outputs[i].Dir, err = ParseProjectDir(output.Dir, projectDir)
		if err != nil {
			return nil, fmt.Errorf("failed to parse output dir: %w", err)
		}

		if !filepath.IsAbs(config.Outputs[i].Dir) {
			config.Outputs[i].Dir = filepath.Join(projectDir, config.Outputs[i].Dir)
		}

		if config.Outputs[i].GoModule != "" && config.Outputs[i].GoPackage == "" {
			goModuleArr := strings.Split(config.Outputs[i].GoModule, "/")
			config.Outputs[i].GoPackage = goModuleArr[len(goModuleArr)-1]
		}
	}

	config.__filepath__ = configPath

	return &config, nil
}
