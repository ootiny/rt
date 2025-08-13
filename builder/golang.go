package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var goEngineMap = map[string]string{
	"net/http": "engine_net_http.go",
}

func toGolangName(name string) string {
	if name[0] >= 'a' && name[0] <= 'z' {
		return strings.ToUpper(name[:1]) + name[1:]
	} else {
		return name
	}
}

func toGolangType(location string, goModule string, currentPackage string, name string) (string, string) {
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
	case "Bytes":
		return "[]byte", ""
	default:
		// if name is List<innter>, then return []inner
		if strings.HasPrefix(name, "List<") && strings.HasSuffix(name, ">") {
			innerType := name[5 : len(name)-1]
			ret, pkg := toGolangType(location, goModule, currentPackage, innerType)
			return fmt.Sprintf("[]%s", ret), pkg
		} else if strings.HasPrefix(name, "Map<") && strings.HasSuffix(name, ">") {
			innerType := name[4 : len(name)-1] // Remove "Map<" and ">"
			ret, pkg := toGolangType(location, goModule, currentPackage, innerType)
			return fmt.Sprintf("map[string]%s", ret), pkg
		} else if strings.HasPrefix(name, DBPrefix) || strings.HasPrefix(name, APIPrefix) {
			nameArr := strings.Split(name, "@")
			if len(nameArr) == 2 {
				pkgName := NamespaceToFolder(location, nameArr[0])

				if pkgName == currentPackage {
					return nameArr[1], ""
				} else {
					pkg := fmt.Sprintf("\t\"%s/%s\"", goModule, pkgName)
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

func (p *GoBuilder) Prepare() error {
	switch p.output.Kind {
	case "server":
		assetEgineFile := fmt.Sprintf("assets/go/%s", goEngineMap[p.output.HttpEngine])
		assetCommonFile := "assets/go/server_common.go"

		if err := os.RemoveAll(p.output.Dir); err != nil {
			return fmt.Errorf("failed to remove system dir: %v", err)
		} else if err := os.MkdirAll(p.output.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create system dir: %v", err)
		} else if engineContent, err := assets.ReadFile(assetEgineFile); err != nil {
			return fmt.Errorf("failed to read assets file: %v", err)
		} else if commonContent, err := assets.ReadFile(assetCommonFile); err != nil {
			return fmt.Errorf("failed to read assets file: %v", err)
		} else {
			// replace package name or namespace if needed
			replaceName := "package _rt_package_name_"
			replaceContent := fmt.Sprintf("package %s", p.output.GoPackage)
			engineContent = []byte(strings.ReplaceAll(string(engineContent), replaceName, replaceContent))
			commonContent = []byte(strings.ReplaceAll(string(commonContent), replaceName, replaceContent))

			if err := WriteGeneratedFile(filepath.Join(p.output.Dir, "engine.go"), string(engineContent)); err != nil {
				return fmt.Errorf("failed to write assets file: %v", err)
			} else if err := WriteGeneratedFile(filepath.Join(p.output.Dir, "common.go"), string(commonContent)); err != nil {
				return fmt.Errorf("failed to write assets file: %v", err)
			} else {
				return nil
			}
		}
	case "client":
		return fmt.Errorf("not implemented")
	default:
		return fmt.Errorf("unknown output kind: %s", p.output.Kind)
	}
}

func (p *GoBuilder) BuildClient() error {
	return fmt.Errorf("not implemented")
}

func (p *GoBuilder) BuildServer() error {
	for _, apiConfig := range p.apiConfigs {
		if err := p.buildServerWithConfig(apiConfig); err != nil {
			return err
		}
	}
	return nil
}

func (p *GoBuilder) buildServerWithConfig(apiConfig APIConfig) error {
	if apiConfig.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	currentPackage := NamespaceToFolder(p.location, apiConfig.Namespace)

	outDir := filepath.Join(p.output.Dir, currentPackage)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	header := fmt.Sprintf("package %s\n", currentPackage)

	imports := []string{}

	defines := []string{}

	actions := []string{}

	registerFuncs := []string{}

	needImportBasePackage := false

	// definitions
	for name, define := range apiConfig.Definitions {
		if len(define.Attributes) > 0 {
			attributes := []string{}
			fullDefineName := apiConfig.Namespace + "@" + name
			for _, attribute := range define.Attributes {
				attrType, pkg := toGolangType(p.location, p.output.GoModule, currentPackage, attribute.Type)
				if pkg != "" {
					imports = append(imports, pkg)
				}

				attributes = append(attributes, fmt.Sprintf(
					"\t%s %s `json:\"%s\" required:\"%t\"`",
					toGolangName(attribute.Name),
					attrType,
					attribute.Name,
					attribute.Required,
				))
			}

			defines = append(defines, fmt.Sprintf(
				"// definition: %s",
				fullDefineName,
			))
			defines = append(defines, fmt.Sprintf(
				"type %s struct {\n%s\n}\n",
				name,
				strings.Join(attributes, "\n"),
			))

			if strings.HasPrefix(apiConfig.Namespace, DBPrefix) {
				defines = append(defines, fmt.Sprintf(
					"type %sBytes = []byte",
					name,
				))
				defines = append(defines, fmt.Sprintf(
					"func Unmarshal%s(data []byte, v *%s) error {\n\t return %s.JsonUnmarshal(data, v)\n}",
					name,
					name,
					p.output.GoPackage,
				))
				defines = append(defines, fmt.Sprintf(
					"func %sBytesTo%s(data []byte) (*%s, error) {\n\tvar v %s\n\tif err := %s.JsonUnmarshal(data, &v); err != nil {\n\t\treturn nil, err\n\t}\n\treturn &v, nil\n}\n",
					name,
					name,
					name,
					name,
					p.output.GoPackage,
				))

				needImportBasePackage = true
			}
		}
	}

	// actions
	if len(apiConfig.Actions) > 0 {
		needImportBasePackage = true

		for name, action := range apiConfig.Actions {
			parameters := []string{}
			structParameters := []string{}
			callParameters := []string{}
			fullActionName := apiConfig.Namespace + ":" + name
			for _, parameter := range action.Parameters {
				typeName, typePkg := toGolangType(p.location, p.output.GoModule, currentPackage, parameter.Type)
				if typePkg != "" {
					imports = append(imports, typePkg)
				}
				parameters = append(parameters, fmt.Sprintf(
					"%s %s",
					parameter.Name,
					typeName,
				))
				goParameterName := toGolangName(parameter.Name)
				structParameters = append(structParameters, fmt.Sprintf(
					"\t\t%s %s `json:\"%s\" required:\"%t\"`",
					goParameterName,
					typeName,
					parameter.Name,
					parameter.Required,
				))
				callParameters = append(callParameters, "v."+goParameterName)
			}

			returnType, typePkg := toGolangType(p.location, p.output.GoModule, currentPackage, action.Return.Type)
			if typePkg != "" {
				imports = append(imports, typePkg)
			}

			returnStr := fmt.Sprintf("%s.Error", p.output.GoPackage)
			if returnType != "" {
				returnStr = fmt.Sprintf("(%s, %s.Error)", returnType, p.output.GoPackage)
			}

			actions = append(actions, fmt.Sprintf(
				"// Action: %s",
				fullActionName,
			))
			actions = append(actions, fmt.Sprintf(
				"var fn%s Func%s",
				name,
				name,
			))
			actions = append(actions, fmt.Sprintf(
				"type Func%s = func(%s) %s",
				name,
				strings.Join(parameters, ", "),
				returnStr,
			))
			actions = append(actions, fmt.Sprintf(
				"func Hook%s (fn Func%s) {\n\tfn%s = fn\n}\n",
				name,
				name,
				name,
			))
			funcBody := ""

			if len(structParameters) > 0 {
				funcBody += fmt.Sprintf("\n\t\tvar v struct {\n\t%s\n\t\t}", strings.Join(structParameters, "\n"))
				funcBody += fmt.Sprintf("\n\t\tif err := %s.JsonUnmarshal(data, &v); err != nil {\n\t\t\treturn nil\n\t\t}\n", p.output.GoPackage)
			}

			funcBody += fmt.Sprintf(
				"\n\t\tif fn%s == nil {\n\t\t\treturn &%s.Return{Code: %s.ErrActionNotImplemented, Message: \"%s is not implemented\"}\n\t\t}",
				name, p.output.GoPackage, p.output.GoPackage, fullActionName,
			)
			if returnType == "" {
				funcBody += fmt.Sprintf(
					" else if err := fn%s(%s); err != nil {\n\t\t\treturn &%s.Return{Code: err.GetCode(), Message: err.Error()}\n\t\t}",
					name, strings.Join(callParameters, ", "), p.output.GoPackage,
				)
				funcBody += fmt.Sprintf(" else {\n\t\t\treturn &%s.Return{}\n\t\t}", p.output.GoPackage)
			} else {
				funcBody += fmt.Sprintf(
					" else if result, err := fn%s(%s); err != nil {\n\t\t\treturn &%s.Return{Code: err.GetCode(), Message: err.Error()}\n\t\t}",
					name, strings.Join(callParameters, ", "), p.output.GoPackage,
				)
				funcBody += fmt.Sprintf(" else {\n\t\t\treturn &%s.Return{Data: result}\n\t\t}", p.output.GoPackage)
			}

			registerFuncs = append(registerFuncs, fmt.Sprintf(
				"\t%s.RegisterHandler(\"%s\", func(ctx %s.Context, response %s.Response, data []byte) *%s.Return {%s\n\t})",
				p.output.GoPackage,
				fullActionName,
				p.output.GoPackage,
				p.output.GoPackage,
				p.output.GoPackage,
				funcBody,
			))
		}
	}

	if needImportBasePackage {
		imports = append(imports, fmt.Sprintf("\t\"%s\"", p.output.GoModule))
	}

	importsContent := ""
	if len(imports) > 0 {
		imports = slices.Compact(imports)
		importsContent = fmt.Sprintf("import (\n%s\n)\n", strings.Join(imports, "\n")) + "\n"
	}

	defineContent := ""
	if len(defines) > 0 {
		defineContent = strings.Join(defines, "\n") + "\n"
	}

	actionContent := ""
	if len(actions) > 0 {
		actionContent = strings.Join(actions, "\n") + "\n"
	}

	registerContent := ""
	if len(registerFuncs) > 0 {
		registerContent = fmt.Sprintf("func init() {\n%s\n}", strings.Join(registerFuncs, "\n"))
	}

	return WriteGeneratedFile(filepath.Join(outDir, "gen.go"), fmt.Sprintf(
		"%s%s%s%s%s",
		header,
		importsContent,
		defineContent,
		actionContent,
		registerContent,
	))
}
