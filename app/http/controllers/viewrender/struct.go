package viewrender

type PageData[T any] struct {
	Title       string
	Description string
	Keywords    string
	PageType    string
	IsDevMode   bool
	Content     T
}
