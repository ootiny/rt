package builder

type DBTableColumn struct {
	Type      string          `json:"type"`
	QueryMap  map[string]bool `json:"queryMap"`
	Unique    bool            `json:"unique"`
	Index     bool            `json:"index"`
	Order     bool            `json:"order"`
	Required  bool            `json:"required"`
	LinkTable string          `json:"linkTable"`
}

type DBTableViewColumn struct {
	Name      string `json:"name"`
	LinkTable string `json:"linkTable"`
	LinkView  string `json:"linkView"`
}

type DBTableView struct {
	Columns       []*DBTableViewColumn `json:"columns"`
	ColumnsSelect string               `json:"columnsSelect"`
	CacheSecond   int64                `json:"cacheSecond"`
}

type DBTable struct {
	Version string                    `json:"version"`
	Name    string                    `json:"name"`
	Columns map[string]*DBTableColumn `json:"columns"`
	Views   map[string]*DBTableView   `json:"views"`
	Hash    string                    `json:"hash"`
	File    string                    `json:"file"`
}
