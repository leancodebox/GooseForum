package component

type Page[T any] struct {
	List  []T   `json:"list"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}
