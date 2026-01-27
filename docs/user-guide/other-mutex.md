> 分布式锁

## 自定义

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

redis实现的分布式互斥锁

```go
import (
	"github.com/asjard/asjard/pkg/xredis"
	"github.com/asjard/asjard/pkg/mutex"
)

func main() {
	var count int64
	do := func() error {
		count += 1
		return nil
	}
	// redis分布式锁
	rediLock, err := xredis.NewLock()
	if err != nil {
		panic(err)
	}
	m := &mutext.Mutex{Locker: redisLock}
	wg := sync.WaitGroup{}
	key := "test_do_with_lock"
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.TryLock(context.Background(), key, do, WithMaxRetries(-1)); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println(count)
	// Output: 10
}
```

mysql实现的分布式互斥锁

```go
import (
	"github.com/asjard/asjard/pkg/xgorm"
	"github.com/asjard/asjard/pkg/mutex"
)

func main() {
	var count int64
	do := func() error {
		count += 1
		return nil
	}
	// gorm分布式锁
	gormLock, err := xgorm.NewLock()
	if err != nil {
		panic(err)
	}
	m := &mutext.Mutex{Locker: gormLock}
	wg := sync.WaitGroup{}
	key := "test_do_with_lock"
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.TryLock(context.Background(), key, do, WithMaxRetries(-1)); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println(count)
	// Output: 10
}
```
