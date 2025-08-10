package builder

import "fmt"

func TypescriptPrepare(output RTOutputConfig) error {
	return nil
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
