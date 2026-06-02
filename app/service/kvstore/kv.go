package kvstore

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
)

var (
	currentMu   sync.RWMutex
	current     *store
	connectOnce = sync.OnceValues(connect)

	// ErrNotFound 键不存在
	ErrNotFound = errors.New("kvstore: key not found")
	// ErrInvalidKey key 为空或不合法
	ErrInvalidKey = errors.New("kvstore: invalid key")
	// ErrClosed 表示 kvstore 已关闭，应用生命周期内不应再次连接。
	ErrClosed = errors.New("kvstore: closed")
)

const maxWriteConflictRetries = 32

// UpdateAction 表示 UpdateBytes 回调希望对当前 key 执行的操作。
type UpdateAction int

const (
	// UpdateKeep 保持当前值不变。
	UpdateKeep UpdateAction = iota
	// UpdateSet 写入回调返回的新值。
	UpdateSet
)

// store 是包内部复用的 Badger-backed KV 连接。
// Badger 推荐在应用生命周期内打开一次并复用；外部调用 Get/Set/UpdateBytes 时会懒加载同一个实例。
type store struct {
	db      *badger.DB
	stopGC  chan struct{}
	gcWg    sync.WaitGroup
	closeMu sync.RWMutex
	closed  bool
}

func connectStore() (*store, error) {
	instance, err := connectOnce()
	if err != nil {
		return nil, err
	}
	if instance.isClosed() {
		return nil, ErrClosed
	}
	return instance, nil
}

func connect() (*store, error) {
	path := preferences.Get("badger.path", "./storage/badger")
	opts := badger.DefaultOptions(path).WithLogger(nil)
	database, err := badger.Open(opts)
	if err != nil {
		slog.Error("kvstore: failed to open database", "path", path, "error", err)
		return nil, fmt.Errorf("kvstore: open database: %w", err)
	}

	instance := &store{
		db:     database,
		stopGC: make(chan struct{}),
	}
	instance.gcWg.Add(1)
	go instance.runGC()

	currentMu.Lock()
	current = instance
	currentMu.Unlock()

	closer.Register(func() error {
		instance.close()
		return nil
	})
	return instance, nil
}

func (s *store) update(fn func(database *badger.DB) error) error {
	var err error
	for attempt := 0; attempt <= maxWriteConflictRetries; attempt++ {
		err = s.withDB(fn)
		if !errors.Is(err, badger.ErrConflict) {
			return err
		}
		// Badger 的读改写事务在并发更新同一 key 时可能冲突，基础层统一短重试。
		time.Sleep(time.Duration(attempt+1) * time.Millisecond)
	}
	return fmt.Errorf("kvstore: write conflict after %d retries: %w", maxWriteConflictRetries, err)
}

func (s *store) withDB(fn func(database *badger.DB) error) error {
	s.closeMu.RLock()
	if s.closed {
		s.closeMu.RUnlock()
		return ErrClosed
	}
	err := fn(s.db)
	s.closeMu.RUnlock()
	return err
}

func (s *store) isClosed() bool {
	s.closeMu.RLock()
	closed := s.closed
	s.closeMu.RUnlock()
	return closed
}

func validKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	return nil
}

func newEntry(key string, value []byte, ttl time.Duration) *badger.Entry {
	entry := badger.NewEntry([]byte(key), append([]byte{}, value...))
	if ttl > 0 {
		entry.WithTTL(ttl)
	}
	return entry
}

func copyItemValue(item *badger.Item) ([]byte, error) {
	var value []byte
	err := item.Value(func(val []byte) error {
		value = append([]byte{}, val...)
		return nil
	})
	return value, err
}

func (s *store) runGC() {
	defer s.gcWg.Done()
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for {
				select {
				case <-s.stopGC:
					slog.Debug("kvstore: stop gc worker")
					return
				default:
				}
				if err := s.db.RunValueLogGC(0.5); err != nil {
					break
				}
			}
		case <-s.stopGC:
			slog.Debug("kvstore: stop gc worker")
			return
		}
	}
}

// Close 关闭当前已打开的 KV 连接；未连接时不做任何操作。
func Close() {
	currentMu.RLock()
	store := current
	currentMu.RUnlock()
	if store == nil {
		return
	}
	store.close()
}

func (s *store) close() {
	s.closeMu.Lock()
	if s.closed {
		s.closeMu.Unlock()
		return
	}
	s.closed = true
	close(s.stopGC)
	s.closeMu.Unlock()

	s.gcWg.Wait()
	if err := s.db.Close(); err != nil {
		slog.Error("kvstore: failed to close database", "error", err)
	}

	currentMu.Lock()
	if current == s {
		current = nil
	}
	currentMu.Unlock()
}

// Set 存储字符串。
func Set(key string, value string, ttl time.Duration) error {
	return SetBytes(key, []byte(value), ttl)
}

// SetBytes 存储字节数组。
func SetBytes(key string, value []byte, ttl time.Duration) error {
	if err := validKey(key); err != nil {
		return err
	}
	instance, err := connectStore()
	if err != nil {
		return err
	}
	return instance.update(func(database *badger.DB) error {
		return database.Update(func(txn *badger.Txn) error {
			return txn.SetEntry(newEntry(key, value, ttl))
		})
	})
}

// Get 获取字符串值。
func Get(key string) (string, error) {
	value, err := GetBytes(key)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// GetBytes 获取原始字节数组。
func GetBytes(key string) ([]byte, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	instance, err := connectStore()
	if err != nil {
		return nil, err
	}
	var value []byte
	err = instance.withDB(func(database *badger.DB) error {
		return database.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					return ErrNotFound
				}
				return err
			}
			value, err = copyItemValue(item)
			return err
		})
	})
	return value, err
}

// UpdateBytes 在单个 Badger 事务内完成一个 key 的读改写。
// updater 可能因事务冲突被重试，调用方应只基于 current/exists 计算返回值，避免在回调里执行外部副作用。
func UpdateBytes(key string, ttl time.Duration, updater func(current []byte, exists bool) (UpdateAction, []byte, error)) error {
	if err := validKey(key); err != nil {
		return err
	}
	instance, err := connectStore()
	if err != nil {
		return err
	}
	return instance.update(func(database *badger.DB) error {
		return database.Update(func(txn *badger.Txn) error {
			var current []byte
			exists := true
			item, err := txn.Get([]byte(key))
			if err != nil {
				if errors.Is(err, badger.ErrKeyNotFound) {
					exists = false
				} else {
					return err
				}
			} else if current, err = copyItemValue(item); err != nil {
				return err
			}

			action, next, err := updater(current, exists)
			if err != nil {
				return err
			}
			switch action {
			case UpdateKeep:
				return nil
			case UpdateSet:
				return txn.SetEntry(newEntry(key, next, ttl))
			default:
				return fmt.Errorf("kvstore: unknown update action %d", action)
			}
		})
	})
}
