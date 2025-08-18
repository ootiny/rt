package builder

import (
	"fmt"
	"strings"
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
