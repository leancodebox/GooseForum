package users

import (
	"github.com/leancodebox/GooseForum/bundles/algorithm"
	"github.com/leancodebox/goose/queryopt"
)

func Get(id any) (entity Entity, err error) {
	err = builder().Where(pid, id).First(&entity).Error
	return
}

func Verify(username string, password string) (*Entity, error) {
	var user Entity
	err := builder().Where(queryopt.Eq(fieldUsername, username)).First(&user).Error
	if err != nil {
		return &user, err
	}
	err = algorithm.VerifyEncryptPassword(user.Password, password)
	if err != nil {
		return &Entity{}, err
	}
	return &user, nil
}

func MakeUser(name string, password string, email string) *Entity {
	user := Entity{Username: name, Email: email}
	user.SetPassword(password)
	return &user
}

func Create(entity *Entity) error {
	return builder().Create(&entity).Error
}

func All() (entities []*Entity) {
	builder().Find(&entities)
	return
}

func GetByUsername(username string) (entities *Entity) {
	builder().Where(queryopt.Eq(fieldUsername, username)).First(entities)
	return
}
