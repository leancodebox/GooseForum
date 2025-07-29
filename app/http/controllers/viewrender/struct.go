package viewrender

type PageMeta struct {
	// 基础SEO
	Title        string
	Description  string
	Keywords     string
	CanonicalURL string

	// OpenGraph
	OGType        string
	OGTitle       string
	OGDescription string
	OGImage       string
	OGURL         string

	Favicon    string
	ThemeColor string

	// 结构化数据
	SchemaOrgJSON string
}
