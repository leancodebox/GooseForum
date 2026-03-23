package kvstore

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

var (
	db     *badger.DB
	once   sync.Once
	stopGC = make(chan struct{})

	// ErrNotFound 键不存在
	ErrNotFound = errors.New("kvstore: key not found")

	onlineUserPrefix = "online:user:"
)

// SetUserOnline 标记用户在线，设置 3 分钟 TTL
// 每次用户活动时调用此方法续期
func SetUserOnline(userId string) error {
	key := onlineUserPrefix + userId
	return SetBytes(key, []byte{1}, 3*time.Minute)
}

// GetOnlineUserCount 获取当前在线用户数
// 通过遍历匹配前缀且未过期的 Key 数量来实现
func GetOnlineUserCount() (int, error) {
	count := 0
	err := getDB().View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // 只需要统计数量，不需要预取 Value
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(onlineUserPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			count++
		}
		return nil
	})
	return count, err
}

// getDB 获取 BadgerDB 实例
func getDB() *badger.DB {
	once.Do(func() {
		path := preferences.Get("badger.path", "./storage/badger")
		opts := badger.DefaultOptions(path).WithLogger(nil)
		var err error
		db, err = badger.Open(opts)
		if err != nil {
			slog.Error("kvstore: failed to open database", "path", path, "error", err)
			panic(err)
		}

		// 启动后台 GC 协程
		go runGC()

		// 自动注册到全局关闭管理器
		closer.Register(func() error {
			Close()
			return nil
		})
	})
	return db
}

// runGC 定期执行 Value Log GC 以回收磁盘空间
func runGC() {
	// 每 10 分钟检查一次是否需要 GC
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// RunValueLogGC 参数 0.5 表示如果一个 log 文件中超过 50% 的数据是过期的，则重写该文件
			// 循环执行直到没有文件可以被 GC
			for {
				if err := db.RunValueLogGC(0.5); err != nil {
					break
				}
			}
		case <-stopGC:
			slog.Debug("kvstore: stop gc worker")
			return
		}
	}
}

// Close 关闭数据库连接并停止后台任务
func Close() {
	if db != nil {
		select {
		case <-stopGC:
			// 已经关闭
		default:
			close(stopGC)
		}
		if err := db.Close(); err != nil {
			slog.Error("kvstore: failed to close database", "error", err)
		}
		db = nil
	}
}

// Set 存储字符串
func Set(key string, value string, ttl time.Duration) error {
	return SetBytes(key, []byte(value), ttl)
}

// SetBytes 存储字节数组
func SetBytes(key string, value []byte, ttl time.Duration) error {
	return getDB().Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value)
		if ttl > 0 {
			entry.WithTTL(ttl)
		}
		return txn.SetEntry(entry)
	})
}

// Get 获取字符串值
func Get(key string) (string, error) {
	val, err := GetBytes(key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// GetBytes 获取原始字节数组
func GetBytes(key string) ([]byte, error) {
	var value []byte
	err := getDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}
			return err
		}
		return item.Value(func(val []byte) error {
			value = append([]byte{}, val...)
			return nil
		})
	})
	return value, err
}

// Exists 检查键是否存在
func Exists(key string) bool {
	err := getDB().View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})
	return err == nil
}

// Increment 增加（或减少）计数器的值并返回新值
// 适用于统计在线人数 (delta 为 1 或 -1) 或帖子浏览量
func Increment(key string, delta int64) (int64, error) {
	var val int64
	err := getDB().Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				val = delta
			} else {
				return err
			}
		} else {
			err = item.Value(func(v []byte) error {
				if len(v) < 8 {
					val = delta
					return nil
				}
				val = int64(binary.BigEndian.Uint64(v)) + delta
				return nil
			})
			if err != nil {
				return err
			}
		}

		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(val))
		return txn.Set([]byte(key), buf)
	})
	return val, err
}

// GetInt64 获取 int64 类型的计数器值
func GetInt64(key string) (int64, error) {
	var val int64
	err := getDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotFound
			}
			return err
		}
		return item.Value(func(v []byte) error {
			if len(v) < 8 {
				return errors.New("kvstore: invalid data length for int64")
			}
			val = int64(binary.BigEndian.Uint64(v))
			return nil
		})
	})
	return val, err
}

// GetManyInt64 批量获取 int64 类型的计数器值
// 适用于列表页展示（如批量获取一页帖子的浏览量）
// 返回结果 map，key 为传入的 key，value 为计数值（不存在的 key 默认为 0）
func GetManyInt64(keys []string) (map[string]int64, error) {
	results := make(map[string]int64, len(keys))
	err := getDB().View(func(txn *badger.Txn) error {
		for _, key := range keys {
			item, err := txn.Get([]byte(key))
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					results[key] = 0
					continue
				}
				return err
			}
			err = item.Value(func(v []byte) error {
				if len(v) >= 8 {
					results[key] = int64(binary.BigEndian.Uint64(v))
				} else {
					results[key] = 0
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return results, err
}

// Delete 删除键
func Delete(key string) error {
	return getDB().Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})
}
