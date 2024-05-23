package datastruct

type Option[K, V any] struct {
	Name  K `json:"name"`
	Label K `json:"label,omitempty"`
	Value V `json:"value"`
}
