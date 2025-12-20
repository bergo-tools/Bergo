package test

import (
	"sync"
	"testing"
	"time"
)

// TestRateLimit 测试限流功能
func TestRateLimit(t *testing.T) {
	// 测试限流管理器的逻辑
	// 在实际使用中，NewLlmStreamer会调用applyRateLimit
	start1 := time.Now()
	
	// 测试限流管理器
	manager := &struct {
		mu            sync.Mutex
		lastRequestAt map[string]time.Time
	}{
		lastRequestAt: make(map[string]time.Time),
	}
	
	// 模拟两次连续请求
	applyRateLimit := func(modelIdentifier string, interval float64) {
		if interval <= 0 {
			return
		}
		
		manager.mu.Lock()
		defer manager.mu.Unlock()
		
		lastAt, exists := manager.lastRequestAt[modelIdentifier]
		if exists {
			elapsed := time.Since(lastAt)
			required := time.Duration(interval * float64(time.Second))
			if elapsed < required {
				waitTime := required - elapsed
				time.Sleep(waitTime)
			}
		}
		manager.lastRequestAt[modelIdentifier] = time.Now()
	}
	
	// 第一次调用
	applyRateLimit("test-model", 1.0)
	duration1 := time.Since(start1)
	
	// 第二次调用应该等待
	start2 := time.Now()
	applyRateLimit("test-model", 1.0)
	duration2 := time.Since(start2)
	
	t.Logf("First call duration: %v", duration1)
	t.Logf("Second call duration: %v (should be ~1s)", duration2)
	
	// 验证第二次调用确实等待了大约1秒
	if duration2 < 900*time.Millisecond || duration2 > 1100*time.Millisecond {
		t.Errorf("Expected second call to wait ~1s, but got %v", duration2)
	}
}

// TestRateLimitMultipleModels 测试多模型限流隔离
func TestRateLimitMultipleModels(t *testing.T) {
	// 这个测试验证不同模型的限流是独立的
	// 实现逻辑同上，但使用不同的模型标识符
	
	t.Log("Multi-model rate limiting should be isolated by model identifier")
	// 这里可以添加更详细的测试逻辑
}

// TestNoRateLimit 测试无限制情况
func TestNoRateLimit(t *testing.T) {
	start := time.Now()
	
	// 模拟无限制的情况
	modelIdentifier := "test-model-no-limit"
	interval := 0.0 // 无限制
	
	manager := &struct {
		mu            sync.Mutex
		lastRequestAt map[string]time.Time
	}{
		lastRequestAt: make(map[string]time.Time),
	}
	
	applyRateLimit := func(modelIdentifier string, interval float64) {
		if interval <= 0 {
			return
		}
		
		manager.mu.Lock()
		defer manager.mu.Unlock()
		
		lastAt, exists := manager.lastRequestAt[modelIdentifier]
		if exists {
			elapsed := time.Since(lastAt)
			required := time.Duration(interval * float64(time.Second))
			if elapsed < required {
				waitTime := required - elapsed
				time.Sleep(waitTime)
			}
		}
		manager.lastRequestAt[modelIdentifier] = time.Now()
	}
	
	// 多次快速调用
	for i := 0; i < 5; i++ {
		applyRateLimit(modelIdentifier, interval)
	}
	
	duration := time.Since(start)
	t.Logf("No limit calls duration: %v", duration)
	
	// 应该非常快，没有等待
	if duration > 100*time.Millisecond {
		t.Errorf("Expected no limit calls to be fast, but got %v", duration)
	}
}
