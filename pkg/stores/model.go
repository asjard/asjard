package stores

import (
	"context"
	"reflect"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"golang.org/x/sync/singleflight"
)

// Model provides common data management patterns (CRUD) combining DB and Cache.
type Model struct {
	// sg (Singleflight Group) ensures that for a specific key, only one concurrent
	// request reaches the database at a time, preventing "Cache Breakdown."
	sg singleflight.Group
}

// Modeler interface defines the metadata required for a struct to use the Model manager.
type Modeler interface {
	// ModelName provides a unique identifier for the model, usually "db_table".
	ModelName() string
}

const (
	// DefaultSingleflightKey is used when no specific key is available for synchronization.
	DefaultSingleflightKey = "default"
)

// GetData handles the "Cache-Aside" read pattern.
// Logic:
// 1. Check if cache is enabled and key is present.
// 2. Attempt to fetch from cache.
// 3. If cache miss: Use singleflight to call the 'get' function (DB).
// 4. On DB success: Update cache.
// 5. On DB failure: Store an empty result in cache (Negative Caching) to prevent DB penetration.
func (m *Model) GetData(ctx context.Context, out any, cache Cacher, get func() (any, error)) (err error) {
	if out == nil {
		logger.Error("GetData out is nil")
		return status.InternalServerError()
	}
	// Bypass logic if caching is not configured or disabled.
	if cache == nil || !cache.Enabled() || cache.Key() == "" {
		result, err := get()
		if err != nil {
			return err
		}
		return m.copy(result, out)
	}

	// Try fetching from the cache provider.
	fromCurrent, err := cache.Get(ctx, cache.Key(), out)
	if err != nil {
		// Cache Miss: Fetch from database using Singleflight to protect the DB.
		result, err, _ := m.sg.Do(cache.Key(), get)
		if err != nil {
			// DB Miss/Error: Cache the empty result for a short period (Negative Caching).
			if rerr := cache.Set(ctx, cache.Key(), out, cache.EmptyExpiresIn()); rerr != nil {
				logger.Error("set empty into cache fail", "err", rerr)
			}
			return err
		}
		// DB Success: Populate cache with the new data.
		if err := cache.Set(ctx, cache.Key(), result, cache.ExpiresIn()); err != nil {
			logger.Error("set cache fail", "key", cache.Key(), "err", err)
		}
		return m.copy(result, out)
	}

	// Optional: Extend the cache TTL if AutoRefresh is enabled.
	if fromCurrent && cache.AutoRefresh() {
		if err := cache.Refresh(ctx, cache.Key(), out, cache.ExpiresIn()); err != nil {
			logger.Error("refresh cache expire fail", "key", cache.Key(), "err", err)
		}
	}
	return nil
}

// SetData handles the "Cache-Aside" write pattern with Delayed Double Delete.
// This pattern is critical for maintaining consistency in distributed environments.
// Logic:
//  1. Immediately delete the cache.
//  2. Execute the database update.
//  3. Wait for a short duration (100ms) and delete the cache again to clear any
//     stale data written by concurrent reads during the DB update.
func (m *Model) SetData(ctx context.Context, set func() error, caches ...Cacher) error {
	// First Delete: Invalidate cache before DB update.
	for _, cache := range caches {
		if err := m.delCache(ctx, cache); err != nil {
			return err
		}
	}
	// DB Update.
	if err := set(); err != nil {
		return err
	}

	// Second Delete (Delayed): Clean up any potential race condition data.
	go func(ctx context.Context, caches ...Cacher) {
		select {
		case <-time.After(100 * time.Millisecond):
			for _, cache := range caches {
				if err := m.delCache(ctx, cache); err != nil {
					logger.L(ctx).Error("delay delete cache fail", "err", err)
				}
			}
		}
	}(ctx, caches...)
	return nil
}

// SetAndGetData updates the database and immediately synchronizes the cache.
// Used for "Write-Through" style operations where cache consistency is a priority.
func (m *Model) SetAndGetData(ctx context.Context, out any, cache Cacher, set func() (any, error)) error {
	if out == nil {
		logger.Error("SetAndGetData out is nil")
		return status.InternalServerError()
	}

	// Invalidate current cache before modification.
	if err := m.delCache(ctx, cache); err != nil {
		return err
	}

	// Execute DB update.
	result, err := set()
	if err != nil {
		return err
	}

	// Update cache with the new DB value.
	if cache != nil && cache.Enabled() {
		if err := cache.Set(ctx, cache.Key(), result, cache.ExpiresIn()); err != nil {
			logger.Error("SetAndGetData set cache fail", "key", cache.Key(), "err", err)
		}
	}
	return m.copy(result, out)
}

// delCache internal helper to safely delete a key from a cache provider.
func (m *Model) delCache(ctx context.Context, cache Cacher) error {
	if cache == nil || !cache.Enabled() {
		return nil
	}
	if err := cache.Del(ctx, cache.Key()); err != nil {
		logger.Error("delete cache fail", "key", cache.Key(), "err", err)
		return status.DeleteCacheFailError()
	}
	return nil
}

// copy uses reflection to deep-copy the database result into the output pointer.
// It ensures type safety between the 'get' function return and the user-provided output.
func (m *Model) copy(from, to any) error {
	fromVal := reflect.ValueOf(from)
	toVal := reflect.ValueOf(to)
	// Validation: 'to' must be a pointer so we can modify its value.
	if toVal.Kind() != reflect.Ptr || toVal.IsNil() {
		logger.Error("out must be a non-nil ptr")
		return status.InternalServerError()
	}
	// Validation: Ensure types match before copying.
	if fromVal.Type() != toVal.Type() {
		logger.Error("type mismatch: get func return type must be same with out type")
		return status.InternalServerError()
	}
	// Set the value of the 'to' pointer to the value held by 'from'.
	toVal.Elem().Set(fromVal.Elem())
	return nil
}
