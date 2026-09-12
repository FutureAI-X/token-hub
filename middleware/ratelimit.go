package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rlBucket 令牌桶
type rlBucket struct {
	tokens   float64
	lastSeen time.Time
}

var (
	rlMu        sync.Mutex
	rlBuckets   = make(map[int]*rlBucket)
	rlLastSweep time.Time
)

// maxRateLimitBuckets 同时跟踪的用户数上限，防止内存无界增长
const maxRateLimitBuckets = 100000

// rateLimitSweepInterval 过期桶清理间隔
const rateLimitSweepInterval = 10 * time.Minute

// RateLimit 按用户维度限流（需在 APIAuth 之后使用，依赖上下文中已设置 user_id）。
// 用于 /v1 业务端点：这些端点每次调用都会真实消耗上游额度，
// 没有限流时单个 Key 即可把上游配额或本地资源打满。
// perMinute 为稳态速率，burst 为允许的瞬时突发量。
func RateLimit(perMinute int, burst int) gin.HandlerFunc {
	if perMinute <= 0 {
		perMinute = 60
	}
	if burst <= 0 {
		burst = 10
	}
	ratePerSecond := float64(perMinute) / 60.0

	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID <= 0 {
			// 未通过 APIAuth 的场景不限流，交由认证中间件处理
			c.Next()
			return
		}

		if !allowRequest(userID, ratePerSecond, float64(burst)) {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message": "请求过于频繁，请稍后再试",
					"type":    "rate_limit_error",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allowRequest 令牌桶判定（线程安全）
func allowRequest(userID int, ratePerSecond, burst float64) bool {
	now := time.Now()

	rlMu.Lock()
	defer rlMu.Unlock()

	sweepRateLimitBucketsLocked(now)

	b := rlBuckets[userID]
	if b == nil {
		if len(rlBuckets) >= maxRateLimitBuckets {
			// 满载时放弃对新用户限流，宁可放宽也不让内存无界增长
			return true
		}
		b = &rlBucket{tokens: burst, lastSeen: now}
		rlBuckets[userID] = b
	}

	// 按经过的时间补充令牌，上限为 burst
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.lastSeen = now
	b.tokens = min(burst, b.tokens+elapsed*ratePerSecond)

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// sweepRateLimitBucketsLocked 清理长期不活跃的桶（调用方须持有 rlMu）
func sweepRateLimitBucketsLocked(now time.Time) {
	if !rlLastSweep.IsZero() && now.Sub(rlLastSweep) < rateLimitSweepInterval {
		return
	}
	rlLastSweep = now

	for id, b := range rlBuckets {
		if now.Sub(b.lastSeen) > rateLimitSweepInterval {
			delete(rlBuckets, id)
		}
	}
}
