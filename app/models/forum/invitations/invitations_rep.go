package invitations

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"gorm.io/gorm"
)

var (
	ErrInvalidParameters  = errors.New("invalid invitation parameters")
	ErrRequired           = errors.New("invitation is required")
	ErrNotFound           = errors.New("invitation not found")
	ErrExpired            = errors.New("invitation expired")
	ErrRevoked            = errors.New("invitation revoked")
	ErrExhausted          = errors.New("invitation exhausted")
	ErrReserved           = errors.New("invitation reserved")
	ErrReservationExpired = errors.New("invitation reservation expired")
)

func Digest(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

func Create(createdBy uint64, maxUses int, expiresAt *time.Time) (Entity, string, error) {
	return CreateWithDB(dbconnect.Connect(), createdBy, maxUses, expiresAt)
}

func CreateWithDB(db *gorm.DB, createdBy uint64, maxUses int, expiresAt *time.Time) (Entity, string, error) {
	if db == nil || createdBy == 0 || maxUses < 1 || maxUses > 100000 {
		return Entity{}, "", ErrInvalidParameters
	}
	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return Entity{}, "", ErrInvalidParameters
	}
	var entropy [24]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return Entity{}, "", err
	}
	raw := "gf_inv_" + base64.RawURLEncoding.EncodeToString(entropy[:])
	entity := Entity{Hash: Digest(raw), Prefix: raw[:16], MaxUses: maxUses, CreatedBy: createdBy, ExpiresAt: expiresAt}
	if err := db.Create(&entity).Error; err != nil {
		return Entity{}, "", err
	}
	return entity, raw, nil
}

func Redeem(raw string, userID uint64) (Entity, error) {
	return RedeemWithDB(dbconnect.Connect(), raw, userID, time.Now())
}

// Reserve creates a short-lived hold without consuming a use. The caller must
// Commit or Release it after the surrounding registration work finishes.
func Reserve(raw string, now time.Time, ttl time.Duration) (Entity, string, error) {
	return ReserveWithDB(dbconnect.Connect(), raw, now, ttl)
}

func ReserveWithDB(db *gorm.DB, raw string, now time.Time, ttl time.Duration) (Entity, string, error) {
	if db == nil || strings.TrimSpace(raw) == "" || ttl <= 0 {
		return Entity{}, "", ErrInvalidParameters
	}
	var entity Entity
	var reservationBytes [24]byte
	if _, err := rand.Read(reservationBytes[:]); err != nil {
		return Entity{}, "", err
	}
	reservation := "gf_res_" + base64.RawURLEncoding.EncodeToString(reservationBytes[:])
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("hash = ?", Digest(raw)).First(&entity).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if entity.RevokedAt != nil {
			return ErrRevoked
		}
		if entity.ExpiresAt != nil && !entity.ExpiresAt.After(now) {
			return ErrExpired
		}
		if entity.UsedCount >= entity.MaxUses {
			return ErrExhausted
		}
		if entity.ReservationUntil != nil && entity.ReservationUntil.After(now) {
			return ErrReserved
		}
		until := now.Add(ttl)
		result := tx.Model(&Entity{}).Where("id = ? AND revoked_at IS NULL AND used_count < max_uses AND (expires_at IS NULL OR expires_at > ?) AND (reservation_until IS NULL OR reservation_until <= ?)", entity.Id, now, now).
			Updates(map[string]any{"reservation_hash": Digest(reservation), "reservation_until": until})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrReserved
		}
		entity.ReservationHash = Digest(reservation)
		entity.ReservationUntil = &until
		return nil
	})
	return entity, reservation, err
}

func Commit(raw, reservation string, userID uint64, now time.Time) (Entity, error) {
	return CommitWithDB(dbconnect.Connect(), raw, reservation, userID, now)
}

func CommitWithDB(db *gorm.DB, raw, reservation string, userID uint64, now time.Time) (Entity, error) {
	if db == nil || strings.TrimSpace(raw) == "" || strings.TrimSpace(reservation) == "" || userID == 0 {
		return Entity{}, ErrInvalidParameters
	}
	var entity Entity
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("hash = ?", Digest(raw)).First(&entity).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if entity.RevokedAt != nil {
			return ErrRevoked
		}
		if entity.ExpiresAt != nil && !entity.ExpiresAt.After(now) {
			return ErrExpired
		}
		if entity.UsedCount >= entity.MaxUses {
			return ErrExhausted
		}
		if entity.ReservationHash != Digest(reservation) {
			if entity.ReservationUntil != nil && !entity.ReservationUntil.After(now) {
				return ErrReservationExpired
			}
			return ErrReserved
		}
		result := tx.Model(&Entity{}).Where("id = ? AND reservation_hash = ? AND reservation_until > ? AND used_count < max_uses", entity.Id, Digest(reservation), now).
			Updates(map[string]any{"used_count": gorm.Expr("used_count + 1"), "last_used_at": now, "last_used_user_id": userID, "reservation_hash": "", "reservation_until": nil})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrReservationExpired
		}
		return tx.First(&entity, entity.Id).Error
	})
	return entity, err
}

func Release(raw, reservation string, now time.Time) error {
	return ReleaseWithDB(dbconnect.Connect(), raw, reservation, now)
}

func ReleaseWithDB(db *gorm.DB, raw, reservation string, now time.Time) error {
	if db == nil || strings.TrimSpace(raw) == "" || strings.TrimSpace(reservation) == "" {
		return ErrInvalidParameters
	}
	return db.Model(&Entity{}).Where("hash = ? AND reservation_hash = ?", Digest(raw), Digest(reservation)).Updates(map[string]any{"reservation_hash": "", "reservation_until": nil}).Error
}

func List() ([]Entity, error) {
	rows := make([]Entity, 0)
	err := dbconnect.Connect().Order("created_at DESC").Find(&rows).Error
	return rows, err
}

type PageQuery struct {
	Page     int
	PageSize int
}

func Page(q PageQuery) (list []Entity, page, pageSize int, total int64, err error) {
	page = max(q.Page, 1)
	pageSize = pageutil.BoundPageSize(q.PageSize)
	db := pageQuery(dbconnect.Connect())
	if err = db.Count(&total).Error; err != nil {
		return nil, page, pageSize, 0, err
	}
	err = db.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error
	return list, page, pageSize, total, err
}

func pageQuery(db *gorm.DB) *gorm.DB {
	return db.Table(tableName + " AS registration_invitations").
		Select("registration_invitations.*, task_queue.status AS email_task_status").
		Joins("LEFT JOIN task_queue ON task_queue.id = registration_invitations.email_task_id")
}

func Revoke(id uint64, now time.Time) error {
	if id == 0 {
		return ErrInvalidParameters
	}
	return dbconnect.Connect().Model(&Entity{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
}

// RedeemWithDB atomically consumes one use and records the user that consumed it.
func RedeemWithDB(db *gorm.DB, raw string, userID uint64, now time.Time) (Entity, error) {
	raw = strings.TrimSpace(raw)
	if db == nil || raw == "" || userID == 0 {
		return Entity{}, ErrInvalidParameters
	}
	var entity Entity
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("hash = ?", Digest(raw)).First(&entity).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if entity.RevokedAt != nil {
			return ErrRevoked
		}
		if entity.ExpiresAt != nil && !entity.ExpiresAt.After(now) {
			return ErrExpired
		}
		if entity.UsedCount >= entity.MaxUses {
			return ErrExhausted
		}
		result := tx.Model(&Entity{}).
			Where("id = ? AND revoked_at IS NULL AND used_count < max_uses AND (expires_at IS NULL OR expires_at > ?)", entity.Id, now).
			Updates(map[string]any{"used_count": gorm.Expr("used_count + 1"), "last_used_at": now, "last_used_user_id": userID})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrExhausted
		}
		return tx.First(&entity, entity.Id).Error
	})
	return entity, err
}
