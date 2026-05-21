package queryopt

import (
	"fmt"
)

type InValue interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint16 | ~uint32 | ~uint64 |
		~string
}

func NotIn[T InValue](field string, list []T) (string, any) {
	return fmt.Sprintf(`%v not in ?`, field), list
}

func In[T InValue](field string, list []T) (string, any) {
	return fmt.Sprintf(`%v in ?`, field), list
}

func Gt(field, value any) (string, any) {
	return fmt.Sprintf(`%v > ?`, field), value
}

func Ge(field, value any) (string, any) {
	return fmt.Sprintf(`%v >= ?`, field), value
}

func Lt(field, value any) (string, any) {
	return fmt.Sprintf(`%v < ?`, field), value
}

func Le(field, value any) (string, any) {
	return fmt.Sprintf(`%v <= ?`, field), value
}

func Eq(field, value any) (string, any) {
	return fmt.Sprintf(`%v = ?`, field), value
}

func Like(field, value string) (string, string) {
	return fmt.Sprintf(`%v like ?`, field), fmt.Sprintf(`%%%v%%`, value)
}

func LeftLike(field, value string) (string, string) {
	return fmt.Sprintf(`%v like ?`, field), fmt.Sprintf(`%%%v`, value)
}

func RightLike(field, value string) (string, string) {
	return fmt.Sprintf(`%v like ?`, field), fmt.Sprintf(`%v%%`, value)
}

func Desc(field string) string {
	return fmt.Sprintf(`%v desc`, field)
}

func Asc(field string) string {
	return fmt.Sprintf(`%v asc`, field)
}

func Ne(field, value any) (string, any) {
	return fmt.Sprintf(`%v <> ?`, field), value
}

func IsNull(field string) string {
	return fmt.Sprintf(`%v IS NULL`, field)
}

func IsNotNull(field string) string {
	return fmt.Sprintf(`%v IS NOT NULL`, field)
}
