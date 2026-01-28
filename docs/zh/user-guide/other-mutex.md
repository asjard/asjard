> 分布式锁

## 自定义需实现的方法

```go
// Locker 互斥锁需要实现的方法
type Locker interface {
	// Lock 加锁
	Lock(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
	// Unlock 解锁
	Unlock(ctx context.Context, key, threadId string) bool
	// 续期
	KeepAlive(ctx context.Context, key, threadId string, expiresIn time.Duration) bool
}
```

## 使用

```go
import "github.com/asjard/asjard/pkg/mutex"

m := &mutext.Mutex{Locker: &customeLock{}}
m.TryLock(context.Background(), key, do, mutex.WithMaxRetries(-1))
```
