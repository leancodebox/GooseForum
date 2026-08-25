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
	GetPending(context.Context, string) (*Metadata, error)
	MarkReady(context.Context, string) (*Metadata, error)
	Delete(context.Context, string) error
}

type Service struct {
	writeStore Store
	stores     map[string]Store
	repository metadataRepository
}

func New(writeStore Store, readStores ...Store) (*Service, error) {
	return newService(writeStore, databaseMetadataRepository{}, readStores...)
}

func newService(writeStore Store, repository metadataRepository, readStores ...Store) (*Service, error) {
	if isNil(writeStore) {
		return nil, errors.New("file storage store is required")
	}
	if isNil(repository) {
		return nil, errors.New("file storage metadata repository is required")
	}
	stores := make(map[string]Store, len(readStores)+1)
	for _, store := range append([]Store{writeStore}, readStores...) {
		if isNil(store) || store.Driver() == "" {
			return nil, errors.New("file storage driver is required")
		}
		if _, exists := stores[store.Driver()]; exists {
			return nil, fmt.Errorf("file storage driver %q is configured more than once", store.Driver())
		}
		stores[store.Driver()] = store
	}
	return &Service{writeStore: writeStore, stores: stores, repository: repository}, nil
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
	if err := validatePutRequest(request); err != nil {
		return nil, err
	}

	metadata, err := service.repository.Create(ctx, Metadata{
		Name:          request.Name,
		ContentType:   request.ContentType,
		Size:          request.Size,
		UserId:        request.UserId,
		StorageDriver: service.writeStore.Driver(),
	})
	if err != nil {
		return nil, err
	}
	writeRequest := WriteRequest{Name: request.Name, ContentType: request.ContentType, Size: request.Size, Body: request.Body}
	if err := service.writeStore.Put(ctx, writeRequest); err != nil {
		return nil, service.rollbackPut(ctx, request.Name, err)
	}
	metadata, err = service.repository.MarkReady(ctx, request.Name)
	if err != nil {
		return nil, service.rollbackPut(ctx, request.Name, err)
	}
	return metadata, nil
}

func validatePutRequest(request PutRequest) error {
	if err := validateMetadataRequest(request.Name, request.ContentType, request.Size); err != nil {
		return err
	}
	if request.Body == nil {
		return errors.New("file storage body is required")
	}
	return nil
}

func validateMetadataRequest(name, contentType string, size int64) error {
	if name == "" {
		return errors.New("file storage name is required")
	}
	if contentType == "" {
		return errors.New("file storage content type is required")
	}
	if size < 0 {
		return errors.New("file storage size cannot be negative")
	}
	return nil
}

func (service *Service) rollbackPut(ctx context.Context, name string, cause error) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	return errors.Join(cause, service.writeStore.Delete(cleanupCtx, name), service.repository.Delete(cleanupCtx, name))
}

func (service *Service) Open(ctx context.Context, name string) (*Object, error) {
	metadata, err := service.Stat(ctx, name)
	if err != nil {
		return nil, err
	}
	store, err := service.storeFor(*metadata)
	if err != nil {
		return nil, err
	}
	stored, err := store.Open(ctx, name)
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
	store, err := service.storeFor(*metadata)
	if err != nil {
		return err
	}
	if err := store.Delete(ctx, name); err != nil {
		return err
	}
	return service.repository.Delete(ctx, name)
}

func (service *Service) storeFor(metadata Metadata) (Store, error) {
	store, ok := service.stores[metadata.StorageDriver]
	if !ok {
		return nil, fmt.Errorf("file storage driver %q is not configured", metadata.StorageDriver)
	}
	return store, nil
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
func Configure(writeStore Store, readStores ...Store) error {
	service, err := New(writeStore, readStores...)
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
