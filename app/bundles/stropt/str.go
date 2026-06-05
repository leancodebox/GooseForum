// Package stropt provides string case and pluralization helpers.
package stropt

import (
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

// Plural returns the English plural form of word.
func Plural(word string) string {
	return pluralize.NewClient().Plural(word)
}

// Singular returns the English singular form of word.
func Singular(word string) string {
	return pluralize.NewClient().Singular(word)
}

// Snake converts s to snake_case.
func Snake(s string) string {
	return strcase.ToSnake(s)
}

// Camel converts s to CamelCase.
func Camel(s string) string {
	return strcase.ToCamel(s)
}

// LowerCamel converts s to lowerCamelCase.
func LowerCamel(s string) string {
	return strcase.ToLowerCamel(s)
}
