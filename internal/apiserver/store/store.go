package store

import (
	"context"

	"github.com/costa92/go-protoc/v2/pkg/store/where"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*datastore)))

// 移除全局单例变量，避免内存泄漏和并发问题

type IStore interface {
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
	// TX is used to implement transactions in the Biz layer.
	TX(ctx context.Context, fn func(ctx context.Context) error) error
}

// transactionKey is the key used to store transaction context in context.Context.
type transactionKey struct{}

// datastore is the concrete implementation of the IStore.
type datastore struct {
	core *gorm.DB

	// Additional database instances can be added as needed.
	// Example: fake *gorm.DB
}

// Ensure datastore implements the IStore.
var _ IStore = (*datastore)(nil)

// NewStore creates a new datastore instance.
// 优化: 移除单例模式，每次创建新实例，通过依赖注入管理生命周期
func NewStore(db *gorm.DB) *datastore {
	return &datastore{core: db}
}

// DB filters the database instance based on the input conditions (wheres).
// If no conditions are provided, the function returns the database instance
// from the context (transaction instance or core database instance).
func (store *datastore) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	db := store.core
	// Attempt to retrieve the transaction instance from the context.
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
		db = tx
	}

	// Apply each provided 'where' condition to the query.
	for _, whr := range wheres {
		db = whr.Where(db)
	}
	return db
}

// FakeDB is used to demonstrate multiple database instances.
// It returns a nil gorm.DB, indicating a fake database.
func (ds *datastore) FakeDB(ctx context.Context) *gorm.DB { return nil }

// TX starts a new transaction instance.
// nolint: fatcontext
func (store *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return store.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}
