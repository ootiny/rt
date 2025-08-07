package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func toGolangFolderByNamespace(location string, namespace string) string {
	//  change all namespace to lowercase
	namespace = strings.ToLower(namespace)

	// replace . with _
	namespace = strings.ReplaceAll(namespace, ".", "_")

	if location == MainLocation {
		return namespace
	} else {
		return location + "_" + namespace
	}
}

func toGolangName(name string) string {
	if name[0] >= 'a' && name[0] <= 'z' {
		return strings.ToUpper(name[:1]) + name[1:]
	} else {
		return name
	}
}

func toGolangType(goModule string, location string, name string) (string, string) {
	name = strings.TrimSpace(name)

	switch name {
	case "String":
		return "string", ""
	case "Float64":
		return "float64", ""
	case "Int64":
		return "int64", ""
	case "Bool":
		return "bool", ""
	case "Byte":
		return "byte", ""
	case "Bytes":
		return "[]byte", ""
	case "Error":
		return "error", ""
	default:
		// if name is List<innter>, then return []inner
		if strings.HasPrefix(name, "List<") && strings.HasSuffix(name, ">") {
			innerType := name[5 : len(name)-1]
			ret, pkg := toGolangType(goModule, location, innerType)
			return fmt.Sprintf("[]%s", ret), pkg
		} else if strings.HasPrefix(name, "Map<") && strings.HasSuffix(name, ">") {
			innerType := name[4 : len(name)-1] // Remove "Map<" and ">"
			ret, pkg := toGolangType(goModule, location, innerType)
			return fmt.Sprintf("map[string]%s", ret), pkg
		} else if strings.HasPrefix(name, "DB.") || strings.HasPrefix(name, "API.") {
			nameArr := strings.Split(name, "@")
			if len(nameArr) == 2 {
				pkgName := toGolangFolderByNamespace(location, nameArr[0])
				pkg := fmt.Sprintf(
					"\t\"%s/%s/%s\"",
					goModule,
					location,
					pkgName,
				)

				return fmt.Sprintf("%s.%s", pkgName, nameArr[1]), pkg
			} else {
				return name, ""
			}
		} else {
			return name, ""
		}
	}
}

type GoBuilder struct {
	BuildContext
}

func (p *GoBuilder) Build() error {
	if p.apiConfig.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	nsFolder := toGolangFolderByNamespace(p.location, p.apiConfig.Namespace)

	outDir := filepath.Join(p.output.Dir, p.location, nsFolder)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	header := fmt.Sprintf(`// %s: %s
package %s
`, BuilderStartTag, BuilderDescription, nsFolder)

	imports := []string{}

	defines := []string{}

	actions := []string{}

	for name, define := range p.apiConfig.Definitions {
		if len(define.Attributes) > 0 {
			attributes := []string{}
			for _, attribute := range define.Attributes {
				attrType, pkg := toGolangType(p.output.GoModule, p.location, attribute.Type)
				if pkg != "" {
					imports = append(imports, pkg)
				}

				if attribute.Required {
					attributes = append(attributes, fmt.Sprintf(
						"\t%s %s `json:\"%s\" required:\"true\"`",
						toGolangName(attribute.Name),
						attrType,
						attribute.Name,
					))
				} else {
					attributes = append(attributes, fmt.Sprintf(
						"\t%s %s `json:\"%s\"`",
						toGolangName(attribute.Name),
						attrType,
						attribute.Name,
					))
				}
			}

			defines = append(defines, fmt.Sprintf(
				"type %s struct {\n%s\n}\n", name, strings.Join(attributes, "\n")))
		}
	}

	// // TODO: create actions
	// for name, action := range p.apiConfig.Actions {
	// 	parameters := []string{}
	// 	for _, parameter := range action.Parameters {
	// 		parameters = append(parameters, fmt.Sprintf(
	// 			"%s %s",
	// 			parameter.Name,
	// 			toGolangType(parameter.Type),
	// 		))
	// 	}

	// 	returns := []string{
	// 		toGolangType(action.Return.Type),
	// 		"Error",
	// 	}

	// 	actions = append(actions, fmt.Sprintf(
	// 		"type Func%s = func(%s) (%s)",
	// 		name,
	// 		strings.Join(parameters, ", "),
	// 		strings.Join(returns, ", "),
	// 	))
	// 	actions = append(actions, fmt.Sprintf(
	// 		"type Hook%s = func(fn Func%s) error\n",
	// 		name,
	// 		name,
	// 	))
	// }

	importsContent := ""
	if len(imports) > 0 {
		importsContent = fmt.Sprintf("import (\n%s\n)", strings.Join(imports, "\n")) + "\n"
	}

	defineContent := ""
	if len(defines) > 0 {
		defineContent = strings.Join(defines, "\n")
	}

	registerContent := ""
	// registerContent := fmt.Sprintf(
	// 	"func init() {\n\t__system.RegisterHandler(\"%s\", func(w __system.IResponse, r __system.IRequest) {\n\n\t})\n}",
	// 	p.apiConfig.Namespace,
	// )

	content := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n//%s",
		header,
		importsContent,
		defineContent,
		strings.Join(actions, "\n"),
		registerContent,
		BuilderEndTag,
	)

	return os.WriteFile(filepath.Join(outDir, "rt.go"), []byte(content), 0644)
}

func (p *GoBuilder) BuildDB(folder string) error {
	return nil
}
