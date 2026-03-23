package users

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/datacache"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/samber/lo"
)

var userRoleCache = datacache.Cache[uint64]{}

func Get(id any) (entity EntityComplete, err error) {
	err = builder().Where(pid, id).First(&entity).Error
	return
}

func GetRoleId(userId uint64) (roleId uint64, err error) {
	key := fmt.Sprintf("user_role:%d", userId)
	return userRoleCache.GetOrLoadE(key, func() (uint64, error) {
		var entity EntityComplete
		err = builder().Select(fieldRoleId).Where(pid, userId).First(&entity).Error
		return entity.RoleId, err
	}, 30*time.Minute)
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

func GetMaxId() uint64 {
	var entity EntityComplete
	builder().Order(queryopt.Desc(pid)).Limit(1).First(&entity)
	return entity.Id
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
	return lo.KeyBy(GetByIds(userIds), func(v *EntityComplete) uint64 {
		return v.Id
	})
}

// ExistUsername 检查用户名是否已存在
func ExistUsername(username string) bool {
	var id uint64
	return builder().Select("1").Where("username = ?", username).Limit(1).Scan(&id).RowsAffected > 0
}

// ExistEmail 检查邮箱是否已存在
func ExistEmail(email string) bool {
	var id uint64
	return builder().Select("1").Where("email = ?", email).Limit(1).Scan(&id).RowsAffected > 0
}

// GetAll 用于全量导出/修复数据，支持分页查询
func GetAll(offset, limit int) ([]*EntityComplete, error) {
	var entities []*EntityComplete
	err := builder().Offset(offset).Limit(limit).Order("id ASC").Find(&entities).Error
	return entities, err
}

// GetCountGroupByDay 按天统计注册人数
func GetCountGroupByDay() ([]map[string]any, error) {
	var results []map[string]any
	err := builder().Select("DATE(created_at) as date, count(*) as count").Group("date").Order("date ASC").Find(&results).Error
	return results, err
}

func IncrementPrestige(addNumber int64, userId uint64) int64 {
	result := builder().Exec("UPDATE users SET prestige = prestige+? where id = ?", addNumber, userId)
	return result.RowsAffected
}

func QueryById(startId uint64, limit int) (entities []*EntityComplete) {
	builder().Where(queryopt.Gt(pid, startId)).Limit(limit).Order(queryopt.Asc(pid)).Find(&entities)
	return
}
