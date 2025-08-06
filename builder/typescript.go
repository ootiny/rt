package builder

type TypescriptBuilder struct {
	BuildContext
}

func (p *TypescriptBuilder) BuildAPIServer() error {
	return nil
}

func (p *TypescriptBuilder) BuildAPIClient() error {
	return nil
}

func (p *TypescriptBuilder) BuildDBServer() error {
	return nil
}

func (p *TypescriptBuilder) BuildDBClient() error {
	return nil
}
