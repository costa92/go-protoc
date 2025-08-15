package db

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TestMetricsConfig 测试配置系统
func TestMetricsConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := NewMetricsConfig()

		// 验证默认值
		if config.Enabled {
			t.Error("metrics should be disabled by default")
		}

		if config.CollectInterval != 30*time.Second {
			t.Errorf("expected default interval 30s, got %v", config.CollectInterval)
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		config := NewMetricsConfig()

		// 测试验证
		if err := config.Validate(); err != nil {
			t.Errorf("default config should be valid, got error: %v", err)
		}

		// 测试无效配置
		config.CollectInterval = -1 * time.Second
		if err := config.Validate(); err == nil {
			t.Error("negative interval should be invalid")
		}
	})

	t.Run("ConfigOptions", func(t *testing.T) {
		config := NewMetricsConfig()

		// 测试选项应用
		opts := []MetricsOption{
			WithEnabled(true),
			WithCollectInterval(60 * time.Second),
		}

		config.ApplyOptions(opts...)

		if !config.Enabled {
			t.Error("config should be enabled after applying options")
		}

		if config.CollectInterval != 60*time.Second {
			t.Errorf("expected interval 60s, got %v", config.CollectInterval)
		}
	})

	t.Run("ConfigClone", func(t *testing.T) {
		config := NewMetricsConfig()
		config.Enabled = true
		config.CollectInterval = 60 * time.Second

		cloned := config.Clone()
		if cloned.Enabled != config.Enabled {
			t.Error("cloned config should have same enabled state")
		}

		if cloned.CollectInterval != config.CollectInterval {
			t.Error("cloned config should have same interval")
		}
	})
}

// TestPoolMetrics 测试指标收集
func TestPoolMetrics(t *testing.T) {
	t.Run("BasicFunctionality", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPoolMetrics(registry)
		if metrics == nil {
			t.Fatal("NewPoolMetrics() returned nil")
		}

		// 测试启用/禁用
		if metrics.IsEnabled() {
			t.Error("metrics should be disabled initially")
		}

		metrics.Enable()
		if !metrics.IsEnabled() {
			t.Error("metrics should be enabled after Enable()")
		}

		metrics.Disable()
		if metrics.IsEnabled() {
			t.Error("metrics should be disabled after Disable()")
		}
	})

	t.Run("LabelSet", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		metrics := NewPoolMetrics(registry)

		// 测试标签集合
		labelSet := metrics.getLabelSet()
		if labelSet == nil {
			t.Fatal("getLabelSet() returned nil")
		}

		// 测试设置标签
		labelSet.Set("key1", "value1").Set("key2", "value2")
		labels := labelSet.Get()

		if labels["key1"] != "value1" {
			t.Error("label key1 should be set to value1")
		}

		if labels["key2"] != "value2" {
			t.Error("label key2 should be set to value2")
		}

		// 测试释放
		labelSet.Release()
		
		// 标签应该被清空
		if len(labelSet.labels) != 0 {
			t.Error("labels should be cleared after release")
		}
	})
}

// TestCollector 测试收集器
func TestCollector(t *testing.T) {
	t.Run("BasicLifecycle", func(t *testing.T) {
		config := NewMetricsConfig()
		config.Registry = prometheus.NewRegistry()
		config.Enabled = false // 禁用实际收集防止测试中启动goroutine

		collector := NewCollector(config)
		if collector == nil {
			t.Fatal("NewCollector() returned nil")
		}

		// 测试初始状态
		if collector.IsRunning() {
			t.Error("collector should not be running initially")
		}

		// 测试诊断信息
		diagnostics := collector.GetDiagnostics()
		if diagnostics == nil {
			t.Error("GetDiagnostics() should not return nil")
		}

		if diagnostics["state"] != "stopped" {
			t.Errorf("expected state 'stopped', got %v", diagnostics["state"])
		}
	})

	t.Run("BatchSize", func(t *testing.T) {
		config := NewMetricsConfig()
		config.Registry = prometheus.NewRegistry()

		collector := NewCollector(config)

		// 测试设置批处理大小
		collector.SetBatchSize(20)

		// 验证内部状态（通过诊断信息间接验证）
		diagnostics := collector.GetDiagnostics()
		if diagnostics["connection_count"].(int) != 0 {
			t.Error("should have no connections initially")
		}
	})
}

// TestOptimizedCollector 测试优化收集器
func TestOptimizedCollector(t *testing.T) {
	t.Run("BasicFunctionality", func(t *testing.T) {
		config := NewMetricsConfig()
		config.Registry = prometheus.NewRegistry()
		config.Enabled = false // 禁用实际收集

		collector := NewOptimizedCollector(config)
		if collector == nil {
			t.Fatal("NewOptimizedCollector() returned nil")
		}

		// 测试状态
		status := collector.GetStatus()
		if status == nil {
			t.Error("GetStatus() should not return nil")
		}

		if status["running"].(bool) {
			t.Error("collector should not be running initially")
		}
	})

	t.Run("GlobalInstance", func(t *testing.T) {
		collector := GetGlobalOptimizedCollector()
		if collector == nil {
			t.Fatal("GetGlobalOptimizedCollector() returned nil")
		}

		// 多次调用应该返回同一个实例
		collector2 := GetGlobalOptimizedCollector()
		if collector != collector2 {
			t.Error("GetGlobalOptimizedCollector() should return the same instance")
		}
	})
}

// TestInitializationManager 测试初始化管理器
func TestInitializationManager(t *testing.T) {
	t.Run("BasicFunctionality", func(t *testing.T) {
		config := NewMetricsConfig()
		config.Registry = prometheus.NewRegistry()

		manager := NewInitializationManager(config)
		if manager == nil {
			t.Fatal("NewInitializationManager() returned nil")
		}

		// 测试初始状态
		if manager.IsInitialized() {
			t.Error("manager should not be initialized initially")
		}

		// 测试状态
		status := manager.GetStatus()
		if status == nil {
			t.Error("GetStatus() should not return nil")
		}

		if status["initialized"].(bool) {
			t.Error("manager should not be initialized initially")
		}
	})

	t.Run("GlobalInstance", func(t *testing.T) {
		manager := GetGlobalInitManager()
		if manager == nil {
			t.Fatal("GetGlobalInitManager() returned nil")
		}

		// 多次调用应该返回同一个实例
		manager2 := GetGlobalInitManager()
		if manager != manager2 {
			t.Error("GetGlobalInitManager() should return the same instance")
		}
	})
}

// TestQuickStart 测试快速启动
func TestQuickStart(t *testing.T) {
	t.Run("QuickStartOptions", func(t *testing.T) {
		// 测试选项创建（不实际启动）
		opts := []MetricsOption{
			WithEnabled(true),
			WithCollectInterval(60 * time.Second),
		}

		config := NewMetricsConfig()
		config.ApplyOptions(opts...)

		if !config.Enabled {
			t.Error("config should be enabled after applying options")
		}

		if config.CollectInterval != 60*time.Second {
			t.Errorf("expected interval 60s, got %v", config.CollectInterval)
		}
	})
}

// TestMetricsCache 测试缓存功能
func TestMetricsCache(t *testing.T) {
	cache := NewMetricsCache(5 * time.Minute)

	// 测试设置和检查
	cache.Set("test_key", "test_value")
	if !cache.IsValid("test_key") {
		t.Error("cache should be valid for recently set key")
	}

	// 测试不存在的键
	if cache.IsValid("non_existent_key") {
		t.Error("cache should not be valid for non-existent key")
	}

	// 测试大小
	if cache.Size() != 1 {
		t.Errorf("expected cache size 1, got %d", cache.Size())
	}

	// 测试前缀删除
	cache.Set("prefix:key1", "value1")
	cache.Set("prefix:key2", "value2")
	cache.Set("other:key", "value")

	cache.DeletePrefix("prefix:")

	if cache.IsValid("prefix:key1") {
		t.Error("keys with prefix should be deleted")
	}

	if !cache.IsValid("other:key") {
		t.Error("keys without prefix should remain")
	}
}