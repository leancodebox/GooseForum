// Package queryopt builds small GORM-compatible query fragments.
package queryopt

import (
	"fmt"
)

// InValue is the set of primitive values accepted by IN helpers.
type InValue interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint16 | ~uint32 | ~uint64 |
		~string
}

// NotIn returns a "field not in ?" query fragment.
func NotIn[T InValue](field string, list []T) (string, any) {
	return fmt.Sprintf(`%v not in ?`, field), list
}

// In returns a "field in ?" query fragment.
func In[T InValue](field string, list []T) (string, any) {
	return fmt.Sprintf(`%v in ?`, field), list
}

// Gt returns a greater-than query fragment.
func Gt(field, value any) (string, any) {
	return fmt.Sprintf(`%v > ?`, field), value
}

// Ge returns a greater-than-or-equal query fragment.
func Ge(field, value any) (string, any) {
	return fmt.Sprintf(`%v >= ?`, field), value
}

// Lt returns a less-than query fragment.
func Lt(field, value any) (string, any) {
	return fmt.Sprintf(`%v < ?`, field), value
}

// Le returns a less-than-or-equal query fragment.
func Le(field, value any) (string, any) {
	return fmt.Sprintf(`%v <= ?`, field), value
}

// Eq returns an equality query fragment.
func Eq(field, value any) (string, any) {
	return fmt.Sprintf(`%v = ?`, field), value
}

// Like returns a contains LIKE query fragment.
func Like(field, value string) (string, string) {
	return fmt.Sprintf(`%v like ?`, field), fmt.Sprintf(`%%%v%%`, value)
}

// LeftLike returns a suffix LIKE query fragment.
func LeftLike(field, value string) (string, string) {
	return fmt.Sprintf(`%v like ?`, field), fmt.Sprintf(`%%%v`, value)
}

// RightLike returns a prefix LIKE query fragment.
func RightLike(field, value string) (string, string) {
	return fmt.Sprintf(`%v like ?`, field), fmt.Sprintf(`%v%%`, value)
}

// Desc returns a descending order expression.
func Desc(field string) string {
	return fmt.Sprintf(`%v desc`, field)
}

// Asc returns an ascending order expression.
func Asc(field string) string {
	return fmt.Sprintf(`%v asc`, field)
}

// Ne returns a not-equal query fragment.
func Ne(field, value any) (string, any) {
	return fmt.Sprintf(`%v <> ?`, field), value
}

// IsNull returns an IS NULL query expression.
func IsNull(field string) string {
	return fmt.Sprintf(`%v IS NULL`, field)
}

// IsNotNull returns an IS NOT NULL query expression.
func IsNotNull(field string) string {
	return fmt.Sprintf(`%v IS NOT NULL`, field)
}
