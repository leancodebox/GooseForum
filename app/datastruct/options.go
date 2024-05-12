package datastruct

type Option[K, V any] struct {
	Name  K `json:"name"`
	Value V `json:"value"`
}
