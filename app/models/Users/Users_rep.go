package Users

import (
	"github.com/leancodebox/goose/luckrand"
	"github.com/leancodebox/goose/querymaker"
)

func Get(id any) (entity Users, err error) {
	err = builder().Where(pid, id).First(&entity).Error
	return
}

func Verify(username string, password string) (*Users, error) {
	var user Users
	err := builder().Where(querymaker.Eq(fieldUsername, username)).First(&user).Error
	if err != nil {
		return &user, err
	}
	err = luckrand.VerifyPassword(user.Password, password)
	if err != nil {
		return &Users{}, err
	}
	return &user, nil
}

func MakeUser(name string, password string, email string) *Users {
	user := Users{Username: name, Email: email}
	user.SetPassword(password)
	return &user
}

func Create(entity *Users) error {
	return builder().Create(&entity).Error
}

func All() (entities []*Users) {
	builder().Find(&entities)
	return
}

func GetByUsername(username string) (entities *Users) {
	builder().Where(querymaker.Eq(fieldUsername, username)).First(entities)
	return
}
