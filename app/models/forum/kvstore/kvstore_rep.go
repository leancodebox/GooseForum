package kvstore

import (
	"errors"
	queryopt "github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"sync"
	"time"
)

// Set 设置键值对，可选设置过期时间
func Set(key string, value string, ttl ...time.Duration) error {
	kv := Entity{
		Key:   key,
		Value: value,
	}

	if len(ttl) > 0 {
		expiresAt := time.Now().Add(ttl[0])
		kv.ExpiresAt = &expiresAt
		kv.TTL = int(ttl[0].Seconds())
	}

	return builder().Save(&kv).Error
}

// Get 获取值
func Get(key string) (string, error) {
	var kv Entity
	err := builder().Where("key = ? AND (expires_at IS NULL OR expires_at > ?)",
		key, time.Now()).First(&kv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // 键不存在返回空字符串
		}
		return "", err
	}
	return kv.Value, nil
}

// Delete 删除键
func Delete(key string) error {
	return builder().Where("key = ?", key).Delete(&Entity{}).Error
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	var count int64
	err := builder().Where("key = ? AND (expires_at IS NULL OR expires_at > ?)",
		key, time.Now()).Count(&count).Error
	return count > 0, err
}

// TTL 获取剩余生存时间(秒)
func GetTTL(key string) (int64, error) {
	var kv Entity
	err := builder().Where("key = ?", key).First(&kv).Error
	if err != nil {
		return -2, err // -2表示键不存在
	}

	if kv.ExpiresAt == nil || kv.ExpiresAt.IsZero() {
		return -1, nil // -1表示永不过期
	}

	remaining := time.Until(*kv.ExpiresAt).Seconds()
	if remaining < 0 {
		return -2, nil // 已过期
	}

	return int64(remaining), nil
}

// CleanExpired 清理过期键
func CleanExpired() error {
	return builder().Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Delete(&Entity{}).Error
}

var incrementLock sync.Mutex

func Increment(key string, delta int64) (int64, error) {
	incrementLock.Lock()
	defer incrementLock.Unlock()
	var kv Entity
	builder().Where(queryopt.Eq(pid, key)).First(&kv)
	if kv.Key == "" {
		kv.Key = key
		kv.Value = cast.ToString(delta)
		err := builder().Create(&kv).Error
		return delta, err
	} else {
		newValue := cast.ToInt64(kv.Value) + delta
		tx := builder().Exec("UPDATE kv_store SET value = ? where key = ? and value = ? ",
			cast.ToString(newValue), kv.Key, kv.Value)
		if tx.Error != nil {
			return newValue, tx.Error
		}
		if tx.RowsAffected == 0 {
			return newValue, errors.New("IncrementV2 Fail")
		}
		return newValue, nil
	}
}
