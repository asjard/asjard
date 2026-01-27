## redis实现的分布式互斥锁

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
