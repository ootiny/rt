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

func toGolangType(goModule string, location string, currentPackage string, name string) (string, string) {
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
			ret, pkg := toGolangType(goModule, location, currentPackage, innerType)
			return fmt.Sprintf("[]%s", ret), pkg
		} else if strings.HasPrefix(name, "Map<") && strings.HasSuffix(name, ">") {
			innerType := name[4 : len(name)-1] // Remove "Map<" and ">"
			ret, pkg := toGolangType(goModule, location, currentPackage, innerType)
			return fmt.Sprintf("map[string]%s", ret), pkg
		} else if strings.HasPrefix(name, "DB.") || strings.HasPrefix(name, "API.") {
			nameArr := strings.Split(name, "@")
			if len(nameArr) == 2 {
				pkgName := toGolangFolderByNamespace(location, nameArr[0])

				if pkgName == currentPackage {
					return nameArr[1], ""
				} else {
					pkg := fmt.Sprintf(
						"\t\"%s/%s/%s\"",
						goModule,
						location,
						pkgName,
					)

					return fmt.Sprintf("%s.%s", pkgName, nameArr[1]), pkg
				}
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

	currentPackage := toGolangFolderByNamespace(p.location, p.apiConfig.Namespace)

	outDir := filepath.Join(p.output.Dir, p.location, currentPackage)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	header := fmt.Sprintf(`// %s: %s
package %s
`, BuilderStartTag, BuilderDescription, currentPackage)

	imports := []string{}

	defines := []string{}

	actions := []string{}

	registerFuncs := []string{}

	for name, define := range p.apiConfig.Definitions {
		if len(define.Attributes) > 0 {
			attributes := []string{}
			for _, attribute := range define.Attributes {
				attrType, pkg := toGolangType(p.output.GoModule, p.location, currentPackage, attribute.Type)
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
				"type %s struct {\n%s\n}\n",
				name,
				strings.Join(attributes, "\n"),
			))
		}
	}

	if len(p.apiConfig.Actions) > 0 {
		imports = append(imports, fmt.Sprintf(
			"\t\"%s/%s/_rt_system_\"",
			p.output.GoModule,
			p.location,
		))

		for name, action := range p.apiConfig.Actions {
			parameters := []string{}
			for _, parameter := range action.Parameters {
				typeName, typePkg := toGolangType(p.output.GoModule, p.location, currentPackage, parameter.Type)
				if typePkg != "" {
					imports = append(imports, typePkg)
				}
				parameters = append(parameters, fmt.Sprintf(
					"%s %s",
					parameter.Name,
					typeName,
				))
			}

			returnType, typePkg := toGolangType(p.output.GoModule, p.location, currentPackage, action.Return.Type)
			if typePkg != "" {
				imports = append(imports, typePkg)
			}
			returns := []string{
				returnType,
				"error",
			}

			actions = append(actions, fmt.Sprintf(
				"type Func%s = func(%s) (%s)",
				name,
				strings.Join(parameters, ", "),
				strings.Join(returns, ", "),
			))
			actions = append(actions, fmt.Sprintf(
				"type Hook%s = func(fn Func%s) error\n",
				name,
				name,
			))
			registerFuncs = append(registerFuncs, fmt.Sprintf(
				"\t_rt_system_.RegisterHandler(\"%s\", func(ctx *_rt_system_.Context) *_rt_system_.Return {\n\t\treturn nil\n\t})",
				p.apiConfig.Namespace+":"+name,
			))
		}
	}

	importsContent := ""
	if len(imports) > 0 {
		importsContent = fmt.Sprintf("import (\n%s\n)", strings.Join(imports, "\n")) + "\n"
	}

	defineContent := ""
	if len(defines) > 0 {
		defineContent = strings.Join(defines, "\n")
	}

	registerContent := ""
	if len(registerFuncs) > 0 {
		registerContent = fmt.Sprintf("func init() {\n%s\n}", strings.Join(registerFuncs, "\n"))
	}

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
