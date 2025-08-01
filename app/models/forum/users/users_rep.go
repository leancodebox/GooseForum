package users

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"math/rand"
	"time"
)

func Get(id any) (entity EntityComplete, err error) {
	err = builder().Where(pid, id).First(&entity).Error
	return
}

func Verify(usernameOrEmail string, password string) (*EntityComplete, error) {
	var user EntityComplete
	// 尝试通过用户名或邮箱查找用户
	err := builder().Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error
	if err != nil {
		return &user, err
	}
	err = algorithm.VerifyEncryptPassword(user.Password, password)
	if err != nil {
		return &EntityComplete{}, err
	}
	return &user, nil
}

// GetByEmail 通过邮箱获取用户
func GetByEmail(email string) (entity EntityComplete, err error) {
	err = builder().Where("email = ?", email).First(&entity).Error
	return
}

func MakeUser(name string, password string, email string) *EntityComplete {
	user := EntityComplete{Username: name, Email: email}
	user.SetPassword(password)
	user.AvatarUrl = RandAvatarUrl()
	return &user
}

func RandAvatarUrl() string {
	randomNum := rand.Intn(8) + 1
	return fmt.Sprintf("/static/pic/%d.webp", randomNum)
}

func Create(entity *EntityComplete) error {
	return builder().Create(&entity).Error
}

func Save(entity *EntityComplete) error {
	result := builder().Save(entity)
	return result.Error
}

func All() (entities []*EntityComplete) {
	builder().Find(&entities)
	return
}
func GetCount() int64 {
	var count int64
	builder().Count(&count)
	return count
}

func GetMonthCount() int64 {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var count int64
	builder().Where(queryopt.Ge(fieldCreatedAt, firstOfMonth)).Count(&count)
	return count
}

func GetMaxId() uint64 {
	var entity EntityComplete
	builder().Order(queryopt.Desc(pid)).Limit(1).First(&entity)
	return entity.Id
}

func GetByUsername(username string) (entities *EntityComplete) {
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
	Data     []EntityComplete
} {
	var list []EntityComplete
	q.Page = max(q.Page-1, 0)
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
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
		Data     []EntityComplete
	}{Page: q.Page, PageSize: q.PageSize, Data: list, Total: total}
}

func GetByIds(userIds []uint64) (entities []*EntityComplete) {
	if len(userIds) == 0 {
		return
	}
	builder().Where(queryopt.In(pid, userIds)).Find(&entities)
	return
}

func GetMapByIds(userIds []uint64) map[uint64]*EntityComplete {
	return collectionopt.Slice2Map(GetByIds(userIds), func(v *EntityComplete) uint64 {
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

func QueryById(startId uint64, limit int) (entities []*EntityComplete) {
	builder().Where(queryopt.Gt(pid, startId)).Limit(limit).Order(queryopt.Asc(pid)).Find(&entities)
	return
}
