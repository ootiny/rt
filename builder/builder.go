package builder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type IBuilder interface {
	Prepare() error
	BuildServer() error
	BuildClient() error
}

type BuildContext struct {
	location   string
	rtConfig   *RTConfig
	apiConfigs []*APIConfig
	dbConfigs  []*TableConfig
	output     *RTOutputConfig
}

func ParseProjectDir(filePath string, projectDir string) (string, error) {
	// Check for project directory placeholders in the filePath
	patterns := []string{
		"$projectdir",
		"$projectDir",
		"${ProjectDir}",
		"$ProjectDir",
		"$project",
		"$Project",
		"${projectDir}",
		"${projectdir}",
		"${Project}",
		"${project}",
	}

	result := filePath

	for _, pattern := range patterns {
		if strings.HasPrefix(filePath, pattern) {
			result = strings.Replace(result, pattern, projectDir, 1)
			break
		}
	}

	// If result is a relative path, convert it to an absolute path
	if !filepath.IsAbs(result) {
		if absPath, err := filepath.Abs(result); err != nil {
			return "", fmt.Errorf("failed to convert path to absolute path: %w", err)
		} else {
			result = absPath
		}
	}

	return result, nil
}

func Output() error {
	rtConfig, err := LoadRtConfig()
	if err != nil {
		log.Panicf("Failed to load config: %v", err)
	}

	projectDir := filepath.Dir(rtConfig.GetFilePath())
	log.Printf("rt: project dir: %s\n", projectDir)
	log.Printf("rt: config file: %s\n", rtConfig.GetFilePath())

	apiConfigs := []*APIConfig{}
	dbConfigs := []*TableConfig{}

	for _, output := range rtConfig.Outputs {
		walkErr := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			var header struct {
				Version string `json:"version"`
			}

			switch filepath.Ext(path) {
			case ".json", ".yaml", ".yml":
				if err := UnmarshalConfig(path, &header); err != nil {
					return nil // Not a rt config file, just ignore.  continue walking
				} else if slices.Contains(APIVersions, header.Version) {
					if apiConfig, err := LoadAPIConfig(path); err != nil {
						return err
					} else {
						apiConfigs = append(apiConfigs, apiConfig)
						return nil
					}
				} else if slices.Contains(DBVersions, header.Version) {
					if dbConfig, err := LoadTableConfig(path); err != nil {
						return err
					} else {
						dbConfigs = append(dbConfigs, dbConfig)
						return nil
					}
				} else {
					return nil
				}
			default:
				return nil
			}
		})

		if walkErr != nil {
			return fmt.Errorf("error walking project directory: %w", walkErr)
		}

		var builder IBuilder

		context := BuildContext{
			location:   MainLocation,
			rtConfig:   rtConfig,
			apiConfigs: apiConfigs,
			dbConfigs:  dbConfigs,
			output:     output,
		}

		switch output.Language {
		case "go":
			builder = &GoBuilder{
				BuildContext: context,
			}
		case "typescript":
			builder = &TypescriptBuilder{
				BuildContext: context,
			}
		default:
			return fmt.Errorf("unsupported language: %s", context.output.Language)
		}

		if err := builder.Prepare(); err != nil {
			return err
		}

		switch context.output.Kind {
		case "server":
			if err := builder.BuildServer(); err != nil {
				return err
			}
		case "client":
			if err := builder.BuildClient(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported kind: %s", context.output.Kind)
		}
	}

	return nil
}
