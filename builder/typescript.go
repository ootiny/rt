package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func toTypeScriptType(location string, currentPackage string, name string) (string, string) {
	name = strings.TrimSpace(name)

	switch name {
	case "String":
		return "string", ""
	case "Float64":
		return "number", ""
	case "Int64":
		return "number", ""
	case "Bool":
		return "boolean", ""
	case "Bytes":
		return "string", ""
	default:
		// if name is List<innter>, then return []inner
		if strings.HasPrefix(name, "List<") && strings.HasSuffix(name, ">") {
			innerType := name[5 : len(name)-1]
			ret, pkg := toTypeScriptType(location, currentPackage, innerType)
			return fmt.Sprintf("%s[]", ret), pkg
		} else if strings.HasPrefix(name, "Map<") && strings.HasSuffix(name, ">") {
			innerType := name[4 : len(name)-1] // Remove "Map<" and ">"
			ret, pkg := toTypeScriptType(location, currentPackage, innerType)
			return fmt.Sprintf("{ [key: string]: %s }", ret), pkg
		} else if strings.HasPrefix(name, DBPrefix) || strings.HasPrefix(name, APIPrefix) {
			nameArr := strings.Split(name, "@")
			if len(nameArr) == 2 {
				pkgName := NamespaceToFolder(location, nameArr[0])

				if pkgName == currentPackage {
					return nameArr[1], ""
				} else {
					pkg := fmt.Sprintf("import * as %s from \"../%s\"", pkgName, pkgName)
					return pkgName + "." + nameArr[1], pkg
				}
			} else {
				return name, ""
			}
		} else {
			return name, ""
		}
	}
}

func TypescriptPrepare(output RTOutputConfig) error {
	switch output.Kind {
	case "server":
		return fmt.Errorf("not implemented")
	case "client":
		systemDir := filepath.Join(output.Dir, "system")
		if err := os.RemoveAll(output.Dir); err != nil {
			return fmt.Errorf("failed to remove system dir: %v", err)
		} else if err := os.MkdirAll(systemDir, 0755); err != nil {
			return fmt.Errorf("failed to create system dir: %v", err)
		} else if engineContent, err := assets.ReadFile("assets/typescript/utils.ts"); err != nil {
			return fmt.Errorf("failed to read assets file: %v", err)
		} else if err := os.WriteFile(filepath.Join(systemDir, "utils.ts"), engineContent, 0644); err != nil {
			return fmt.Errorf("failed to write assets file: %v", err)
		} else {
			return nil
		}
	default:
		return fmt.Errorf("unknown output kind: %s", output.Kind)
	}
}

type TypescriptBuilder struct {
	BuildContext
}

func (p *TypescriptBuilder) BuildServer() error {
	return fmt.Errorf("not implemented")
}

func (p *TypescriptBuilder) BuildClient() error {
	if p.apiConfig.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	currentPackage := NamespaceToFolder(p.location, p.apiConfig.Namespace)

	outDir := filepath.Join(p.output.Dir, currentPackage)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	imports := []string{}

	defines := []string{}

	actions := []string{}

	// needImportFetchJson := false

	// definitions
	for name, define := range p.apiConfig.Definitions {
		if len(define.Attributes) > 0 {
			attributes := []string{}
			fullDefineName := p.apiConfig.Namespace + "@" + name
			for _, attribute := range define.Attributes {
				attrType, pkg := toTypeScriptType(p.location, currentPackage, attribute.Type)
				if pkg != "" {
					imports = append(imports, pkg)
				}

				attributes = append(attributes, fmt.Sprintf(
					"  %s: %s;",
					attribute.Name,
					attrType,
				))
			}

			defines = append(defines, fmt.Sprintf(
				"// definition: %s",
				fullDefineName,
			))
			defines = append(defines, fmt.Sprintf(
				"export interface %s {\n%s\n}\n",
				name,
				strings.Join(attributes, "\n"),
			))

		}
	}

	importsContent := ""
	if len(imports) > 0 {
		importsContent = strings.Join(imports, "\n") + "\n"
	}

	defineContent := ""
	if len(defines) > 0 {
		defineContent = strings.Join(defines, "\n")
	}

	actionContent := ""
	if len(actions) > 0 {
		actionContent = strings.Join(actions, "\n")
	}

	content := fmt.Sprintf(
		"// %s: %s\n%s\n%s\n%s\n// %s",
		BuilderStartTag,
		BuilderDescription,
		importsContent,
		defineContent,
		actionContent,
		BuilderEndTag,
	)

	return os.WriteFile(filepath.Join(outDir, "index.ts"), []byte(content), 0644)
}
