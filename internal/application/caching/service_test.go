package caching

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCacheClient is a mock implementation of the cache client
type MockCacheClient struct {
	mock.Mock
}

func (m *MockCacheClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheClient) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	args := m.Called(ctx, pattern)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheClient) FlushAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheClient) LogStats() {}

// MockLogger for caching tests
type MockCachingLogger struct {
	mock.Mock
}

func (m *MockCachingLogger) Info(args ...interface{})                  { m.Called(args) }
func (m *MockCachingLogger) Infof(format string, args ...interface{})  { m.Called(format, args) }
func (m *MockCachingLogger) Debug(args ...interface{})                 { m.Called(args) }
func (m *MockCachingLogger) Debugf(format string, args ...interface{}) { m.Called(format, args) }
func (m *MockCachingLogger) Warn(args ...interface{})                  { m.Called(args) }
func (m *MockCachingLogger) Warnf(format string, args ...interface{})  { m.Called(format, args) }
func (m *MockCachingLogger) Error(args ...interface{})                 { m.Called(args) }
func (m *MockCachingLogger) Errorf(err error, format string, args ...interface{}) {
	m.Called(err, format, args)
}
func (m *MockCachingLogger) Fatal(args ...interface{})                 { m.Called(args) }
func (m *MockCachingLogger) Fatalf(format string, args ...interface{}) { m.Called(format, args) }

// Helper function to create test caching service
func setupTestCachingService() (*Service, *MockCacheClient) {
	cacheClient := new(MockCacheClient)
	logger := new(MockCachingLogger)

	// Allow all logger calls
	logger.On("Info", mock.Anything).Maybe()
	logger.On("Infof", mock.Anything, mock.Anything).Maybe()
	logger.On("Debug", mock.Anything).Maybe()
	logger.On("Debugf", mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything).Maybe()
	logger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	logger.On("Error", mock.Anything).Maybe()
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	cfg := config.CacheConfig{
		Enabled:    true,
		DefaultTTL: 3600,
		MaxSize:    1000,
	}

	service := NewService(cacheClient, cfg, logger)
	return service, cacheClient
}

// Test Get - Cache Hit
func TestCacheService_Get_Hit(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	expectedValue := `{"response":"test data"}`
	cacheClient.On("Get", ctx, "test:key").Return(expectedValue, nil)

	// Test
	value, err := service.Get(ctx, "test:key")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
	cacheClient.AssertExpectations(t)
}

// Test Get - Cache Miss
func TestCacheService_Get_Miss(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	cacheClient.On("Get", ctx, "test:key").Return("", errors.New("redis: nil"))

	// Test
	value, err := service.Get(ctx, "test:key")

	// Assertions
	assert.Error(t, err)
	assert.Empty(t, value)
	cacheClient.AssertExpectations(t)
}

// Test Set - Success
func TestCacheService_Set_Success(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	testValue := `{"response":"test data"}`
	testTTL := 1 * time.Hour

	cacheClient.On("Set", ctx, "test:key", testValue, testTTL).Return(nil)

	// Test
	err := service.Set(ctx, "test:key", testValue, testTTL)

	// Assertions
	assert.NoError(t, err)
	cacheClient.AssertExpectations(t)
}

// Test Set - Error
func TestCacheService_Set_Error(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	testValue := `{"response":"test data"}`
	testTTL := 1 * time.Hour

	cacheClient.On("Set", ctx, "test:key", testValue, testTTL).Return(errors.New("redis connection error"))

	// Test
	err := service.Set(ctx, "test:key", testValue, testTTL)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection error")
	cacheClient.AssertExpectations(t)
}

// Test Delete - Success
func TestCacheService_Delete_Success(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	cacheClient.On("Delete", ctx, "test:key").Return(nil)

	// Test
	err := service.Delete(ctx, "test:key")

	// Assertions
	assert.NoError(t, err)
	cacheClient.AssertExpectations(t)
}

// Test Clear - Success
func TestCacheService_Clear_Success(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	cacheClient.On("FlushAll", ctx).Return(nil)

	// Test
	err := service.Clear(ctx)

	// Assertions
	assert.NoError(t, err)
	cacheClient.AssertExpectations(t)
}

// Test InvalidatePattern - Success
func TestCacheService_InvalidatePattern_Success(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	matchingKeys := []string{"model:claude:key1", "model:claude:key2", "model:claude:key3"}

	cacheClient.On("Keys", ctx, "model:claude:*").Return(matchingKeys, nil)
	cacheClient.On("Delete", ctx, "model:claude:key1").Return(nil)
	cacheClient.On("Delete", ctx, "model:claude:key2").Return(nil)
	cacheClient.On("Delete", ctx, "model:claude:key3").Return(nil)

	// Test
	count, err := service.InvalidatePattern(ctx, "model:claude:*")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	cacheClient.AssertExpectations(t)
}

// Test InvalidatePattern - No Matches
func TestCacheService_InvalidatePattern_NoMatches(t *testing.T) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	cacheClient.On("Keys", ctx, "model:nonexistent:*").Return([]string{}, nil)

	// Test
	count, err := service.InvalidatePattern(ctx, "model:nonexistent:*")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	cacheClient.AssertExpectations(t)
}

// Test GenerateCacheKey - Consistency
func TestCacheService_GenerateCacheKey_Consistency(t *testing.T) {
	service, _ := setupTestCachingService()

	model := "claude-3-opus"
	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello, world!"},
	}
	temperature := 0.7
	maxTokens := 1000

	// Generate key multiple times
	key1 := service.GenerateCacheKey(model, messages, temperature, maxTokens)
	key2 := service.GenerateCacheKey(model, messages, temperature, maxTokens)
	key3 := service.GenerateCacheKey(model, messages, temperature, maxTokens)

	// All keys should be identical
	assert.Equal(t, key1, key2)
	assert.Equal(t, key2, key3)
	assert.NotEmpty(t, key1)
}

// Test GenerateCacheKey - Different Inputs
func TestCacheService_GenerateCacheKey_DifferentInputs(t *testing.T) {
	service, _ := setupTestCachingService()

	messages1 := []map[string]interface{}{
		{"role": "user", "content": "Hello!"},
	}
	messages2 := []map[string]interface{}{
		{"role": "user", "content": "Hi!"},
	}

	key1 := service.GenerateCacheKey("claude-3-opus", messages1, 0.7, 1000)
	key2 := service.GenerateCacheKey("claude-3-opus", messages2, 0.7, 1000)
	key3 := service.GenerateCacheKey("claude-3-sonnet", messages1, 0.7, 1000)
	key4 := service.GenerateCacheKey("claude-3-opus", messages1, 0.8, 1000)
	key5 := service.GenerateCacheKey("claude-3-opus", messages1, 0.7, 2000)

	// All keys should be different
	assert.NotEqual(t, key1, key2) // Different messages
	assert.NotEqual(t, key1, key3) // Different model
	assert.NotEqual(t, key1, key4) // Different temperature
	assert.NotEqual(t, key1, key5) // Different max tokens
}

// Test GetStatistics
func TestCacheService_GetStatistics(t *testing.T) {
	service, _ := setupTestCachingService()

	stats := service.GetStatistics()

	// Initial state
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, int64(0), stats.Errors)
}

// Benchmark cache key generation
func BenchmarkGenerateCacheKey(b *testing.B) {
	service, _ := setupTestCachingService()

	messages := []map[string]interface{}{
		{"role": "system", "content": "You are a helpful assistant."},
		{"role": "user", "content": "What is the meaning of life?"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.GenerateCacheKey("claude-3-opus", messages, 0.7, 1000)
	}
}

// Benchmark cache set operation
func BenchmarkCacheSet(b *testing.B) {
	service, cacheClient := setupTestCachingService()
	ctx := context.Background()

	cacheClient.On("Set", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	testValue := `{"response":"test data"}`
	testTTL := 1 * time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.Set(ctx, "test:key", testValue, testTTL)
	}
}
