package stores

import (
	"testing"
	"time"

	"github.com/asjard/asjard/utils"
)

func TestCache(t *testing.T) {
	c := Cache{
		conf: &CacheConfig{
			ExpiresIn:      utils.JSONDuration{Duration: 10 * time.Minute},
			EmptyExpiresIn: utils.JSONDuration{Duration: 5 * time.Minute},
		},
	}
	t.Run("EmptyExpiresIn", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			// [expiresIn, expiresIn*2)
			// 前闭后开
			expiresIn := c.ExpiresIn()
			if expiresIn < c.conf.ExpiresIn.Duration && expiresIn > c.conf.ExpiresIn.Duration*2 {
				t.Error("rand must in [expiresIn, expiresIn*2)")
				t.FailNow()
			}
		}
	})
}
