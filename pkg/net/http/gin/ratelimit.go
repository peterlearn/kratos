package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/peterlearn/kratos/pkg/log"
	limit "github.com/peterlearn/kratos/pkg/ratelimit"
	"github.com/peterlearn/kratos/pkg/ratelimit/bbr"
)

// RateLimiter bbr middleware.
type RateLimiter struct {
	group   *bbr.Group
	logTime int64
}

// NewRateLimiter return a ratelimit middleware.
func NewRateLimiter(conf *bbr.Config) (s *RateLimiter) {
	return &RateLimiter{
		group:   bbr.NewGroup(conf),
		logTime: time.Now().UnixNano(),
	}
}

func (b *RateLimiter) printStats(routePath string, limiter limit.Limiter) {
	now := time.Now().UnixNano()
	if now-atomic.LoadInt64(&b.logTime) > int64(time.Second*3) {
		atomic.StoreInt64(&b.logTime, now)
		log.Info("http.bbr path:%s stat:%+v", routePath, limiter.(*bbr.BBR).Stat())
	}
}

// Limit return a bm handler func.
func (b *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := fmt.Sprintf("%s://%s%s", c.Request.URL.Scheme, c.Request.Host, c.Request.URL.Path)
		limiter := b.group.Get(uri)
		done, err := limiter.Allow(c)
		if err != nil {
			_metricServerBBR.Inc(uri, c.Request.Method)
			res := TOJSON(nil, err)
			c.JSON(http.StatusOK, res)
			c.Abort()
			return
		}
		defer func() {
			done(limit.DoneInfo{Op: limit.Success})
			b.printStats(uri, limiter)
		}()
		c.Next()
	}
}
