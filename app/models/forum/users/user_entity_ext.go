package users

import "github.com/leancodebox/GooseForum/bundles/algorithm"

func (itself *Entity) SetPassword(password string) *Entity {
	itself.Password, _ = algorithm.MakePassword(password)
	return itself
}
