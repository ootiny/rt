package builder

type DBColumnConfig struct {
	Type      string   `json:"type"`
	Query     []string `json:"query"`
	Unique    bool     `json:"unique"`
	Index     bool     `json:"index"`
	Order     bool     `json:"order"`
	LinkTable string   `json:"linkTable"`
	mapQuery  map[string]bool
}

func (p *DBColumnConfig) CanQuery(column string) bool {
	return p.mapQuery[column]
}

type DBViewColumnConfig struct {
	Name      string `json:"name"`
	LinkTable string `json:"linkTable"`
	LinkView  string `json:"linkView"`
}

type DBViewConfig struct {
	Cache       string   `json:"cache"`
	Columns     []string `json:"columns"`
	querySelect string
	cacheSecond int64
}

func (p *DBViewConfig) GetQuerySelect() string {
	return p.querySelect
}

func (p *DBViewConfig) GetCacheSecond() int64 {
	return p.cacheSecond
}

type DBServiceConfig struct {
	Version string                     `json:"version"`
	Table   string                     `json:"table"`
	Columns map[string]*DBColumnConfig `json:"columns"`
	Views   map[string]*DBViewConfig   `json:"views"`
}
