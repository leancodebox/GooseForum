package users

import (
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/queryopt"
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

func Save(entity *Entity) error {
	result := builder().Save(entity)
	return result.Error
}

func All() (entities []*Entity) {
	builder().Find(&entities)
	return
}

func GetByUsername(username string) (entities *Entity) {
	builder().Where(queryopt.Eq(fieldUsername, username)).First(entities)
	return
}

type PageQuery struct {
	Page, PageSize int
	Username       string
	UserId         uint64
	Email          string
}

func Page(q PageQuery) struct {
	Page     int
	PageSize int
	Total    int64
	Data     []Entity
} {
	var list []Entity
	if q.Page > 0 {
		q.Page -= 1
	} else {
		q.Page = 0
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}
	b := builder()
	cB := builder()
	if q.Username != "" {
		b.Where(queryopt.Like(fieldUsername, q.Username))
		cB.Where(queryopt.Like(fieldUsername, q.Username))
	}
	if q.Email != "" {
		b.Where(queryopt.Like(fieldEmail, q.Email))
		cB.Where(queryopt.Like(fieldEmail, q.Email))
	}
	if q.UserId != 0 {
		b.Where(queryopt.Eq(pid, q.UserId))
		cB.Where(queryopt.Eq(pid, q.UserId))
	}
	b.Limit(q.PageSize).Offset(q.PageSize * q.Page).Order(queryopt.Desc(pid)).Find(&list)

	var total int64
	cB.Count(&total)

	return struct {
		Page     int
		PageSize int
		Total    int64
		Data     []Entity
	}{Page: q.Page, PageSize: q.PageSize, Data: list, Total: total}
}

func GetByIds(userIds []uint64) (entities []*Entity) {
	if len(userIds) == 0 {
		return
	}
	builder().Where(queryopt.In(pid, userIds)).Find(&entities)
	return
}

func GetMapByIds(userIds []uint64) map[uint64]*Entity {
	return collectionopt.Slice2Map(GetByIds(userIds), func(v *Entity) uint64 {
		return v.Id
	})
}

// ExistUsername 检查用户名是否已存在
func ExistUsername(username string) bool {
	var count int64
	builder().Where("username = ?", username).Count(&count)
	return count > 0
}

// ExistEmail 检查邮箱是否已存在
func ExistEmail(email string) bool {
	var count int64
	builder().Where("email = ?", email).Count(&count)
	return count > 0
}

func IncrementPrestige(addNumber int64, userId uint64) int64 {
	result := builder().Exec("UPDATE users SET prestige = prestige+? where id = ?", addNumber, userId)
	return result.RowsAffected
}
