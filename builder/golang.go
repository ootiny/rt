package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func toGolangImport(packageName string) string {
	packageName = strings.ReplaceAll(packageName, "/", "_")
	return "__" + packageName
}

func toGolangName(name string) string {
	if name[0] >= 'a' && name[0] <= 'z' {
		return strings.ToUpper(name[:1]) + name[1:]
	} else {
		return name
	}
}

func toGolangType(name string) string {
	name = strings.TrimSpace(name)

	switch name {
	case "String":
		return "string"
	case "Float64":
		return "float64"
	case "Int64":
		return "int64"
	case "Bool":
		return "bool"
	case "Byte":
		return "byte"
	case "Bytes":
		return "[]byte"
	case "Cookies":
		return "map[string]string"
	case "Headers":
		return "map[string]string"
	case "Error":
		return "error"
	default:
		// if name is List<innter>, then return []inner
		if strings.HasPrefix(name, "List<") && strings.HasSuffix(name, ">") {
			innerType := name[5 : len(name)-1] // Remove "List<" and ">"
			return fmt.Sprintf("[]%s", toGolangType(innerType))
		} else if strings.HasPrefix(name, "Map<") && strings.HasSuffix(name, ">") {
			innerType := name[4 : len(name)-1] // Remove "Map<" and ">"
			return fmt.Sprintf("map[string]%s", toGolangType(innerType))
		} else {
			return name
		}
	}
}

type GolangBuilder struct {
	BuildContext
}

func (p *GolangBuilder) BuildServer() error {
	if p.buildConfig.Package == "" {
		return fmt.Errorf("package is required")
	}

	outDir := filepath.Join(p.output.Dir, p.buildConfig.Package)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	header := fmt.Sprintf(`// %s: %s
package %s
`, BuilderStartTag, BuilderDescription, p.buildConfig.Package)

	imports := []string{
		fmt.Sprintf(
			"\t__system \"%s/__gapi_system__\"",
			p.output.GoModule,
		),
	}

	defines := []string{}

	actions := []string{}

	for name, define := range p.buildConfig.Definitions {
		if define.Import != nil {
			if len(define.Attributes) > 0 {
				return fmt.Errorf("%s can not set attributes when imported", name)
			}

			if p.output.GoModule == "" {
				return fmt.Errorf("goModule must be set in %s, when use imported define", p.rootConfigPath)
			}

			imports = append(imports, fmt.Sprintf(
				"\t%s \"%s/%s\"",
				toGolangImport(define.Import.Package),
				p.output.GoModule,
				define.Import.Package,
			))

			defines = append(defines, fmt.Sprintf(
				"type %s = %s.%s\n",
				name,
				toGolangImport(define.Import.Package),
				define.Import.Name,
			))
		} else if len(define.Attributes) > 0 {
			attributes := []string{}
			for _, attribute := range define.Attributes {
				if attribute.Required {
					attributes = append(attributes, fmt.Sprintf(
						"\t%s %s`json:\"%s\" required:\"true\"`",
						toGolangName(attribute.Name),
						toGolangType(attribute.Type),
						attribute.Name,
					))
				} else {
					attributes = append(attributes, fmt.Sprintf(
						"\t%s %s`json:\"%s\"`",
						toGolangName(attribute.Name),
						toGolangType(attribute.Type),
						attribute.Name,
					))
				}
			}

			defines = append(defines, fmt.Sprintf(
				"type %s struct {\n%s\n}\n", name, strings.Join(attributes, "\n")))
		}
	}

	// TODO: create actions
	for name, action := range p.buildConfig.Actions {
		parameters := []string{}
		for _, parameter := range action.Parameters {
			parameters = append(parameters, fmt.Sprintf(
				"%s %s",
				parameter.Name,
				toGolangType(parameter.Type),
			))
		}
		returns := []string{}
		for _, ret := range action.Returns {
			returns = append(returns, toGolangType(ret.Type))
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
	}

	importsContent := ""
	if len(imports) > 0 {
		importsContent = fmt.Sprintf("import (\n%s\n)", strings.Join(imports, "\n")) + "\n"
	}

	defineContent := ""
	if len(defines) > 0 {
		defineContent = strings.Join(defines, "\n")
	}

	registerContent := fmt.Sprintf(
		"func init() {\n\t__system.RegisterHandler(\"%s\", func(w __system.IResponse, r __system.IRequest) {\n\n\t})\n}",
		p.buildConfig.ApiPath,
	)

	content := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n//%s",
		header,
		importsContent,
		defineContent,
		strings.Join(actions, "\n"),
		registerContent,
		BuilderEndTag,
	)

	return os.WriteFile(filepath.Join(outDir, "gapi.go"), []byte(content), 0644)
}

func (p *GolangBuilder) BuildClient() error {
	return nil
}
