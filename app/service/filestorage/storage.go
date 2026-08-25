package filestorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

type Metadata struct {
	Id            uint64
	Name          string
	ContentType   string
	Size          int64
	UserId        uint64
	StorageDriver string
	StorageStatus string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (metadata Metadata) GetAccessPath() string {
	return urlconfig.FilePath(metadata.Name)
}

type PutRequest struct {
	Name        string
	ContentType string
	Size        int64
	UserId      uint64
	Body        io.Reader
}

type WriteRequest struct {
	Name        string
	ContentType string
	Size        int64
	Body        io.Reader
}

type StoredObject struct {
	Size int64
	Body io.ReadCloser
}

type Object struct {
	Metadata Metadata
	Body     io.ReadCloser
}

// Store only owns file content. Forum metadata is maintained by Service.
type Store interface {
	Driver() string
	Put(context.Context, WriteRequest) error
	Open(context.Context, string) (*StoredObject, error)
	Delete(context.Context, string) error
}

type metadataRepository interface {
	Create(context.Context, Metadata) (*Metadata, error)
	Get(context.Context, string) (*Metadata, error)
	MarkReady(context.Context, string) (*Metadata, error)
	Delete(context.Context, string) error
}

type Service struct {
	store      Store
	repository metadataRepository
}

func New(store Store) (*Service, error) {
	return newService(store, databaseMetadataRepository{})
}

func newService(store Store, repository metadataRepository) (*Service, error) {
	if isNil(store) {
		return nil, errors.New("file storage store is required")
	}
	if isNil(repository) {
		return nil, errors.New("file storage metadata repository is required")
	}
	if store.Driver() == "" {
		return nil, errors.New("file storage driver is required")
	}
	return &Service{store: store, repository: repository}, nil
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (service *Service) Put(ctx context.Context, request PutRequest) (*Metadata, error) {
	if request.Name == "" {
		return nil, errors.New("file storage name is required")
	}
	if request.ContentType == "" {
		return nil, errors.New("file storage content type is required")
	}
	if request.Size < 0 {
		return nil, errors.New("file storage size cannot be negative")
	}
	if request.Body == nil {
		return nil, errors.New("file storage body is required")
	}

	metadata, err := service.repository.Create(ctx, Metadata{
		Name:          request.Name,
		ContentType:   request.ContentType,
		Size:          request.Size,
		UserId:        request.UserId,
		StorageDriver: service.store.Driver(),
	})
	if err != nil {
		return nil, err
	}
	writeRequest := WriteRequest{Name: request.Name, ContentType: request.ContentType, Size: request.Size, Body: request.Body}
	if err := service.store.Put(ctx, writeRequest); err != nil {
		return nil, service.rollbackPut(ctx, request.Name, err)
	}
	metadata, err = service.repository.MarkReady(ctx, request.Name)
	if err != nil {
		return nil, service.rollbackPut(ctx, request.Name, err)
	}
	return metadata, nil
}

func (service *Service) rollbackPut(ctx context.Context, name string, cause error) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return errors.Join(cause, service.store.Delete(cleanupCtx, name), service.repository.Delete(cleanupCtx, name))
}

func (service *Service) Open(ctx context.Context, name string) (*Object, error) {
	metadata, err := service.Stat(ctx, name)
	if err != nil {
		return nil, err
	}
	if err := service.ensureDriver(*metadata); err != nil {
		return nil, err
	}
	stored, err := service.store.Open(ctx, name)
	if err != nil {
		return nil, err
	}
	if stored.Size >= 0 {
		metadata.Size = stored.Size
	}
	return &Object{Metadata: *metadata, Body: stored.Body}, nil
}

func (service *Service) Stat(ctx context.Context, name string) (*Metadata, error) {
	if name == "" {
		return nil, errors.New("file storage name is required")
	}
	return service.repository.Get(ctx, name)
}

func (service *Service) Delete(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("file storage name is required")
	}
	metadata, err := service.repository.Get(ctx, name)
	if err != nil {
		return err
	}
	if err := service.ensureDriver(*metadata); err != nil {
		return err
	}
	if err := service.store.Delete(ctx, name); err != nil {
		return err
	}
	return service.repository.Delete(ctx, name)
}

func (service *Service) ensureDriver(metadata Metadata) error {
	if metadata.StorageDriver != "" && metadata.StorageDriver != service.store.Driver() {
		return fmt.Errorf("file storage driver %q is not configured", metadata.StorageDriver)
	}
	return nil
}

var (
	defaultMu      sync.RWMutex
	defaultService = mustNew(NewDatabaseStore())
)

func mustNew(store Store) *Service {
	service, err := New(store)
	if err != nil {
		panic(err)
	}
	return service
}

// Configure changes the process-wide storage backend used by application services.
func Configure(store Store) error {
	service, err := New(store)
	if err != nil {
		return err
	}
	defaultMu.Lock()
	defaultService = service
	defaultMu.Unlock()
	return nil
}

func current() *Service {
	defaultMu.RLock()
	service := defaultService
	defaultMu.RUnlock()
	return service
}

func Put(ctx context.Context, request PutRequest) (*Metadata, error) {
	return current().Put(ctx, request)
}

func Open(ctx context.Context, name string) (*Object, error) {
	return current().Open(ctx, name)
}

func Stat(ctx context.Context, name string) (*Metadata, error) {
	return current().Stat(ctx, name)
}

func Delete(ctx context.Context, name string) error {
	return current().Delete(ctx, name)
}
