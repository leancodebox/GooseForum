package users

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func Get(id any) (entity EntityComplete, err error) {
	err = builder().Where(pid, id).First(&entity).Error
	return
}

func GetIdentity(id any) (identity Identity, err error) {
	err = builder().Model(&EntityComplete{}).Where(pid, id).First(&identity).Error
	return
}

func Verify(usernameOrEmail string, password string) (*EntityComplete, error) {
	usernameOrEmail = identityKey(usernameOrEmail)
	var user EntityComplete
	var err error
	if strings.Contains(usernameOrEmail, "@") {
		err = builder().Where("email = ?", usernameOrEmail).First(&user).Error
	} else {
		err = builder().Where("username = ?", usernameOrEmail).First(&user).Error
	}
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
	err = builder().Where("email = ?", identityKey(email)).First(&entity).Error
	return
}

func GetByUsername(username string) (entity EntityComplete, err error) {
	err = builder().Where("username = ?", identityKey(username)).First(&entity).Error
	return
}

func MakeUser(name string, password string, email string) *EntityComplete {
	user := EntityComplete{Username: name, Email: email}
	user.SetPassword(password)
	user.AvatarUrl = RandAvatarUrl()
	return &user
}

func RandAvatarUrl() string {
	randomNum := rand.Intn(12) + 1
	return fmt.Sprintf("/static/pic/%d.webp", randomNum)
}

func Create(entity *EntityComplete) error {
	err := builder().Create(entity).Error
	return identityWriteError(err, entity.Username, entity.Email, 0)
}

func Save(entity *EntityComplete) error {
	result := builder().Omit("username", "email", "is_activated", "activated_at", "requires_email_verification", "password", "token_version", "role_id", "is_frozen", "restriction_status", "restriction_until", "restriction_reason", "restriction_note").Save(entity)
	return result.Error
}

// UpdatePasswordByVersion changes the password and invalidates existing sessions
// in one conditional database update.
func UpdatePasswordByVersion(userID, version uint64, passwordHash string) (bool, error) {
	result := builder().Where("id = ? AND token_version = ?", userID, version).
		Updates(map[string]any{"password": passwordHash, "token_version": gorm.Expr("token_version + 1")})
	return result.RowsAffected == 1, result.Error
}

func UpdateWornBadgeCode(userID uint64, badgeCode string) error {
	return builder().
		Where(queryopt.Eq(pid, userID)).
		Update("worn_badge_code", badgeCode).Error
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

type PublicPageQuery struct {
	PageSize          int
	BeforeID, AfterID uint64
}

type PublicPageResult struct {
	HasPrevious bool
	HasNext     bool
	Data        []PublicDirectoryUser
}

// PublicDirectoryUser is the intentionally small read model used by the
// public member directory.
type PublicDirectoryUser struct {
	Id            uint64
	Username      string
	Nickname      string
	AvatarUrl     string
	Bio           string
	Prestige      int64
	WornBadgeCode string
	CreatedAt     time.Time
}

func (user *PublicDirectoryUser) GetWebAvatarUrl() string {
	return (&EntityComplete{AvatarUrl: user.AvatarUrl}).GetWebAvatarUrl()
}

// PublicPage returns only the columns needed by the public member directory.
// Controllers still decide the final public payload shape.
func PublicPage(q PublicPageQuery) PublicPageResult {
	q.PageSize = pageutil.BoundPageSize(q.PageSize)
	listQuery := builder().
		Model(&EntityComplete{}).
		Where("deleted_at IS NULL").
		Where("restriction_status = ? OR (restriction_status = '' AND is_frozen = ?) OR restriction_until <= ?", RestrictionNormal, StatusNormal, time.Now())
	ascending := q.AfterID > 0
	if ascending {
		listQuery = listQuery.Where("id > ?", q.AfterID).Order(queryopt.Asc(pid))
	} else {
		if q.BeforeID > 0 {
			listQuery = listQuery.Where("id < ?", q.BeforeID)
		}
		listQuery = listQuery.Order(queryopt.Desc(pid))
	}

	var list []PublicDirectoryUser
	listQuery.
		Limit(q.PageSize + 1).
		Find(&list)

	hasExtra := len(list) > q.PageSize
	if hasExtra {
		list = list[:q.PageSize]
	}
	if ascending {
		for left, right := 0, len(list)-1; left < right; left, right = left+1, right-1 {
			list[left], list[right] = list[right], list[left]
		}
		return PublicPageResult{HasPrevious: hasExtra, HasNext: len(list) > 0, Data: list}
	}
	return PublicPageResult{HasPrevious: q.BeforeID > 0 && len(list) > 0, HasNext: hasExtra, Data: list}
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

func IncrementPrestige(addNumber int64, userId uint64) int64 {
	result := builder().Exec("UPDATE users SET prestige = prestige+? where id = ?", addNumber, userId)
	return result.RowsAffected
}

// UpdatePrivacy changes only privacy fields, without overwriting profile edits.
func UpdatePrivacy(userID uint64, hideTopics, hideFollowing, hideActivity bool) error {
	return builder().Where("id = ?", userID).Updates(map[string]any{
		"hide_activity": hideActivity, "hide_topics": hideTopics, "hide_following": hideFollowing,
	}).Error
}
