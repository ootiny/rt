package builder

import (
	"fmt"
	"os"
	"path/filepath"
)

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
	return nil
}
