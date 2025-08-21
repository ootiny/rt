package builder

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// embed assets
//
//go:embed all:assets
var assets embed.FS
var APIVersions = []string{"rt.api.v1"}
var DBVersions = []string{"rt.db.v1"}

const MainLocation = "main"
const DBPrefix = "DB."
const APIPrefix = "API."

func TimeStringToDuration(timeStr string) (time.Duration, error) {
	s := strings.TrimSpace(timeStr)
	if s == "" {
		return 0, fmt.Errorf("time string is empty")
	}

	// Validate the whole string first: sequences like 1h30m, 2d, 5m, etc.
	full := regexp.MustCompile(`(?i)^\s*(\d+\s*[dhms])+\s*$`)
	if !full.MatchString(s) {
		return 0, fmt.Errorf("invalid time string: %s", timeStr)
	}

	// Extract all number+unit parts and accumulate.
	partRe := regexp.MustCompile(`(?i)(\d+)\s*([dhms])`)
	parts := partRe.FindAllStringSubmatch(s, -1)
	var total time.Duration
	for _, p := range parts {
		n, err := strconv.ParseInt(p[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid time string: %s", timeStr)
		}
		var unit time.Duration
		switch strings.ToLower(p[2]) {
		case "d":
			unit = 24 * time.Hour
		case "h":
			unit = time.Hour
		case "m":
			unit = time.Minute
		case "s":
			unit = time.Second
		default:
			return 0, fmt.Errorf("invalid time string: %s", timeStr)
		}
		total += time.Duration(n) * unit
	}
	return total, nil
}

var DBBaseTypes = []string{
	"PK", "Bool", "Int64", "Float64",
	"String", "String16", "String32", "String64", "String256",
	"List<String>", "Map<String>",
}

var APIBaseTypes = []string{
	"String",
	"Float64",
	"Int64",
	"Bool",
}

// 将DB类型转换为API类型
func DBTypeToApiType(dbType string) (string, error) {
	switch dbType {
	case "PK":
		return "String", nil
	case "Bool":
		return "Bool", nil
	case "Int64":
		return "Int64", nil
	case "Float64":
		return "Float64", nil
	case "String", "String16", "String32", "String64", "String256":
		return "String", nil
	case "List<String>":
		return "List<String>", nil
	case "Map<String>":
		return "Map<String>", nil
	default:
		if strings.HasPrefix(dbType, "List<") && strings.HasSuffix(dbType, ">") {
			innerType := dbType[5 : len(dbType)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				return DBTypeToApiType(innerType)
			} else {
				return "", fmt.Errorf("invalid column type: %s", dbType)
			}
		} else if strings.HasPrefix(dbType, "Map<") && strings.HasSuffix(dbType, ">") {
			innerType := dbType[4 : len(dbType)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				return DBTypeToApiType(innerType)
			} else {
				return "", fmt.Errorf("invalid column type: %s", dbType)
			}
		} else if strings.HasPrefix(dbType, DBPrefix) {
			columnArray := strings.Split(dbType, "@")
			if len(columnArray) == 1 {
				return fmt.Sprintf("%s@Default", columnArray[0]), nil
			} else if len(columnArray) == 2 {
				return fmt.Sprintf("%s@%s", columnArray[0], columnArray[1]), nil
			} else {
				return "", fmt.Errorf("invalid column type: %s", dbType)
			}
		} else {
			return "", fmt.Errorf("invalid column type: %s", dbType)
		}
	}
}

func DBTypeToTableColumn(dbType string) (*DBTableColumn, error) {
	strType := ""
	strTable := ""
	switch dbType {
	case "PK":
		strType = "String"
	case "Bool":
		strType = "Bool"
	case "Int64":
		strType = "Int64"
	case "Float64":
		strType = "Float64"
	case "String", "String16", "String32", "String64", "String256":
		strType = "String"
	case "List<String>":
		strType = "List<String>"
	case "Map<String>":
		strType = "Map<String>"
	default:
		if strings.HasPrefix(dbType, "List<") && strings.HasSuffix(dbType, ">") {
			innerType := dbType[5 : len(dbType)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				strType = fmt.Sprintf("LKList")
				strTable = NamespaceToTableName(innerType)
			} else {
				return nil, fmt.Errorf("invalid column type: %s", dbType)
			}
		} else if strings.HasPrefix(dbType, "Map<") && strings.HasSuffix(dbType, ">") {
			innerType := dbType[4 : len(dbType)-1]
			if strings.HasPrefix(innerType, DBPrefix) {
				strType = fmt.Sprintf("LKMap")
				strTable = NamespaceToTableName(innerType)
			} else {
				return nil, fmt.Errorf("invalid column type: %s", dbType)
			}
		} else if strings.HasPrefix(dbType, DBPrefix) {
			strType = "LK"
			strTable = NamespaceToTableName(dbType)
		} else {
			return nil, fmt.Errorf("invalid column type: %s", dbType)
		}
	}

	return &DBTableColumn{
		Type:      strType,
		QueryMap:  map[string]bool{},
		Unique:    false,
		Index:     false,
		Order:     false,
		Required:  false,
		LinkTable: strTable,
	}, nil
}

func UnmarshalConfig(filePath string, v any) error {
	if content, err := os.ReadFile(filePath); err != nil {
		return err
	} else {
		switch filepath.Ext(filePath) {
		case ".json":
			return json.Unmarshal(content, v)
		case ".yaml", ".yml":
			return yaml.Unmarshal(content, v)
		default:
			return fmt.Errorf("unsupported file extension: %s", filepath.Ext(filePath))
		}
	}
}

func WriteJSONFile(filePath string, v any) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, content, 0644)
}

func WriteGeneratedFile(filePath string, content string) error {
	const BuilderStartTag = "tag-rt-api-builder-start"
	const BuilderEndTag = "tag-rt-api-builder-end"
	const BuilderDescription = "This file is generated by rt-builder, DO NOT EDIT."

	fileContent := fmt.Sprintf(
		"// %s: %s\n%s\n// %s",
		BuilderStartTag,
		BuilderDescription,
		content,
		BuilderEndTag,
	)

	// todo: create dir if not exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, []byte(fileContent), 0644)
}

func NamespaceToFolder(location string, namespace string) string {
	//  change all namespace to lowercase
	namespace = strings.ToLower(namespace)

	// replace . with _
	namespace = strings.ReplaceAll(namespace, ".", "_")

	if location == MainLocation {
		return namespace
	} else {
		return fmt.Sprintf("%s_%s", location, namespace)
	}
}

func NamespaceToTableName(namespace string) string {
	if after, ok := strings.CutPrefix(namespace, DBPrefix); ok {
		namespace = after
	}

	//  change all namespace to lowercase
	namespace = strings.ToLower(namespace)

	// replace . with _
	namespace = strings.ReplaceAll(namespace, ".", "_")

	return namespace
}
