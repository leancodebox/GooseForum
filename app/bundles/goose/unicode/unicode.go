package unicode

import (
	"regexp"
	"strconv"
)

func Decode(str string) string {
	// 正则表达式，用于匹配Unicode编码部分
	re := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)

	// 将Unicode编码转换为中文字符串
	return re.ReplaceAllStringFunc(str, func(m string) string {
		code, _ := strconv.ParseInt(m[2:], 16, 32)
		return string(rune(code))
	})
}
