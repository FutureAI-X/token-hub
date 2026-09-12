package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// newRateLimitContext 构造带指定 user_id 的测试上下文
func newRateLimitContext(userID int) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/images/generations", nil)
	if userID > 0 {
		c.Set("user_id", userID)
	}
	return c
}

// resetRateLimitState 清空限流状态，保证用例之间互不影响
func resetRateLimitState() {
	rlMu.Lock()
	defer rlMu.Unlock()
	rlBuckets = make(map[int]*rlBucket)
	rlLastSweep = time.Time{}
}

// TestRateLimitAllowsBurstThenBlocks 验证令牌桶：突发额度用尽后立即拒绝
func TestRateLimitAllowsBurstThenBlocks(t *testing.T) {
	resetRateLimitState()

	const burst = 5
	mw := RateLimit(60, burst)

	calls := 0
	handler := func(c *gin.Context) { calls++ }

	for i := 0; i < burst; i++ {
		c := newRateLimitContext(1)
		mw(c)
		if c.IsAborted() {
			t.Fatalf("第 %d 次请求在突发额度内，不应被拒绝", i+1)
		}
		handler(c)
	}

	// 突发额度耗尽后，下一次应立即被拒绝
	c := newRateLimitContext(1)
	mw(c)
	if !c.IsAborted() {
		t.Error("超出突发额度后应返回 429")
	}
	if calls != burst {
		t.Errorf("业务处理函数被调用了 %d 次，期望 %d 次", calls, burst)
	}
}

// TestRateLimitIsolatesUsers 验证限流按用户隔离，单个用户不会影响他人
func TestRateLimitIsolatesUsers(t *testing.T) {
	resetRateLimitState()

	mw := RateLimit(60, 2)

	for i := 0; i < 2; i++ {
		c := newRateLimitContext(100)
		mw(c)
		if c.IsAborted() {
			t.Fatalf("用户 100 第 %d 次请求不应被拒绝", i+1)
		}
	}

	// 用户 100 已耗尽额度
	c := newRateLimitContext(100)
	mw(c)
	if !c.IsAborted() {
		t.Error("用户 100 应已被限流")
	}

	// 用户 200 应有独立额度
	c2 := newRateLimitContext(200)
	mw(c2)
	if c2.IsAborted() {
		t.Error("用户 200 的额度不应受到用户 100 的影响")
	}
}

// TestRateLimitSkipsUnauthenticated 验证未认证请求交由认证中间件处理
func TestRateLimitSkipsUnauthenticated(t *testing.T) {
	resetRateLimitState()

	mw := RateLimit(60, 1)
	for i := 0; i < 5; i++ {
		c := newRateLimitContext(0) // user_id 缺失
		mw(c)
		if c.IsAborted() {
			t.Fatal("未认证请求不应由限流中间件拦截")
		}
	}
}

// TestRateLimitConcurrentSafety 并发调用不应触发竞态（配合 -race 运行）
func TestRateLimitConcurrentSafety(t *testing.T) {
	resetRateLimitState()

	mw := RateLimit(6000, 100)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				c := newRateLimitContext(id%5 + 1)
				mw(c)
			}
		}(i)
	}
	wg.Wait()
}

// TestAllowRequestBucketIsCapped 验证令牌不会无限累积超过 burst
func TestAllowRequestBucketIsCapped(t *testing.T) {
	resetRateLimitState()

	// 首次请求创建桶并消耗 1 个令牌
	if !allowRequest(7, 1.0, 3.0) {
		t.Fatal("首次请求应被允许")
	}

	rlMu.Lock()
	b := rlBuckets[7]
	rlMu.Unlock()

	if b == nil {
		t.Fatal("应已创建令牌桶")
	}
	if b.tokens > 3.0 {
		t.Errorf("令牌数 %v 超过 burst 上限 3.0", b.tokens)
	}
}
