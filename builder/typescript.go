package builder

import "fmt"

type TypescriptBuilder struct {
	BuildContext
}

func (p *TypescriptBuilder) BuildServer() error {
	return fmt.Errorf("not implemented")
}

func (p *TypescriptBuilder) BuildClient() error {
	return nil
}
