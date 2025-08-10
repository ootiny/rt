package builder

import "fmt"

func TypescriptPrepare(output RTOutputConfig) error {
	switch output.Kind {
	case "server":
		return fmt.Errorf("not implemented")
	case "client":
		return nil
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
