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

type TypescriptBuilder struct {
	BuildContext
}

func (p *TypescriptBuilder) Prepare() error {
	switch p.output.Kind {
	case "server":
		return fmt.Errorf("not implemented")
	case "client":
		systemDir := filepath.Join(p.output.Dir, "system")
		if err := os.RemoveAll(p.output.Dir); err != nil {
			return fmt.Errorf("failed to remove system dir: %v", err)
		} else if err := os.MkdirAll(systemDir, 0755); err != nil {
			return fmt.Errorf("failed to create system dir: %v", err)
		} else if engineContent, err := assets.ReadFile("assets/typescript/utils.ts"); err != nil {
			return fmt.Errorf("failed to read assets file: %v", err)
		} else if err := WriteGeneratedFile(filepath.Join(systemDir, "utils.ts"), string(engineContent)); err != nil {
			return fmt.Errorf("failed to write assets file: %v", err)
		} else {
			return nil
		}
	default:
		return fmt.Errorf("unknown output kind: %s", p.output.Kind)
	}
}

func (p *TypescriptBuilder) BuildServer() error {
	return fmt.Errorf("not implemented")
}

func (p *TypescriptBuilder) BuildClient() error {
	rootNode := MakeApiConfigTree(p.apiConfigs)
	if rootNode == nil {
		// no api found
		return nil
	}

	if apiNode := rootNode.children["API"]; apiNode != nil {
		if err := p.buildClientWithConfig(apiNode); err != nil {
			return err
		}
	}

	if dbNode := rootNode.children["DB"]; dbNode != nil {
		if err := p.buildClientWithConfig(dbNode); err != nil {
			return err
		}
	}

	return nil
}

func (p *TypescriptBuilder) buildClientWithConfig(node *APIConfigNode) error {
	if node.namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	currentPackage := NamespaceToFolder(p.location, node.namespace)

	imports := []string{}

	defines := []string{}

	actions := []string{}

	// needImportFetchJson := false

	// definitions
	if node.config != nil {
		for name, define := range node.config.Definitions {
			if len(define.Attributes) > 0 {
				attributes := []string{}
				fullDefineName := node.config.Namespace + "@" + name
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
	}

	// actions
	if node.config != nil && len(node.config.Actions) > 0 {
		imports = append(imports, "import { fetchJson } from \"../system/utils\";")
		for name, action := range node.config.Actions {
			if len(action.Parameters) > 0 {
				attributes := []string{}
				dataAttrs := []string{}
				fullActionName := node.config.Namespace + ":" + name
				method := strings.ToUpper(action.Method)
				for _, attribute := range action.Parameters {
					attrType, pkg := toTypeScriptType(p.location, currentPackage, attribute.Type)
					if pkg != "" {
						imports = append(imports, pkg)
					}

					attributes = append(attributes, fmt.Sprintf(
						"%s: %s",
						attribute.Name,
						attrType,
					))
					if attribute.Required {
						dataAttrs = append(dataAttrs, attribute.Name)
					}
				}

				returnType, pkg := toTypeScriptType(p.location, currentPackage, action.Return.Type)
				if pkg != "" {
					imports = append(imports, pkg)
				}
				if returnType == "" {
					returnType = "void"
				}

				actionStr := fmt.Sprintf("\t// action: %s\n", fullActionName)
				actionStr += fmt.Sprintf("\tasync %s(%s): Promise<%s> {\n", name, strings.Join(attributes, ", "), returnType)
				actionStr += fmt.Sprintf("\t\treturn fetchJson(this.url, \"%s\", \"%s\", { %s })\n", fullActionName, method, strings.Join(dataAttrs, ", "))
				actionStr += "\t}\n"

				actions = append(actions, actionStr)
			}
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
		actionContent = "export class __Main__ {\n"
		actionContent += "\tprivate url: string;\n"
		actionContent += "\tconstructor(url: string) {\n"
		actionContent += "\t\tthis.url = url;\n"
		actionContent += "\t}\n"
		actionContent += "\n"
		actionContent += strings.Join(actions, "\n")
		actionContent += "}\n"
	}

	// build children
	for _, child := range node.children {
		if err := p.buildClientWithConfig(child); err != nil {
			return err
		}
	}

	if node.namespace == "DB" || node.namespace == "API" {
		return nil
	} else if strings.HasPrefix(node.namespace, "DB.") && node.config == nil {
		return nil
	} else {
		// write file
		return WriteGeneratedFile(filepath.Join(p.output.Dir, currentPackage, "index.ts"), fmt.Sprintf(
			"%s%s%s",
			importsContent,
			defineContent,
			actionContent,
		))
	}
}
