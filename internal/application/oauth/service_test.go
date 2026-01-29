package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientRepository is a mock implementation of OAuthClientRepository
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Create(ctx context.Context, client *models.OAuthClient) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) GetByID(ctx context.Context, id string) (*models.OAuthClient, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OAuthClient), args.Error(1)
}

func (m *MockClientRepository) GetByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OAuthClient), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, client *models.OAuthClient) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepository) List(ctx context.Context, limit, offset int) ([]*models.OAuthClient, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.OAuthClient), args.Error(1)
}

// MockTokenRepository is a mock implementation of OAuthTokenRepository
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Create(ctx context.Context, token *models.OAuthToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) GetByAccessToken(ctx context.Context, accessToken string) (*models.OAuthToken, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OAuthToken), args.Error(1)
}

func (m *MockTokenRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.OAuthToken, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OAuthToken), args.Error(1)
}

func (m *MockTokenRepository) RevokeByAccessToken(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeByRefreshToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(args ...interface{})                  { m.Called(args) }
func (m *MockLogger) Infof(format string, args ...interface{})  { m.Called(format, args) }
func (m *MockLogger) Debug(args ...interface{})                 { m.Called(args) }
func (m *MockLogger) Debugf(format string, args ...interface{}) { m.Called(format, args) }
func (m *MockLogger) Warn(args ...interface{})                  { m.Called(args) }
func (m *MockLogger) Warnf(format string, args ...interface{})  { m.Called(format, args) }
func (m *MockLogger) Error(args ...interface{})                 { m.Called(args) }
func (m *MockLogger) Errorf(err error, format string, args ...interface{}) {
	m.Called(err, format, args)
}
func (m *MockLogger) Fatal(args ...interface{})                 { m.Called(args) }
func (m *MockLogger) Fatalf(format string, args ...interface{}) { m.Called(format, args) }

// Helper function to create test OAuth service
func setupTestService() (*Service, *MockClientRepository, *MockTokenRepository) {
	clientRepo := new(MockClientRepository)
	tokenRepo := new(MockTokenRepository)
	logger := new(MockLogger)

	// Allow all logger calls
	logger.On("Info", mock.Anything).Maybe()
	logger.On("Infof", mock.Anything, mock.Anything).Maybe()
	logger.On("Debug", mock.Anything).Maybe()
	logger.On("Debugf", mock.Anything, mock.Anything).Maybe()
	logger.On("Warn", mock.Anything).Maybe()
	logger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	logger.On("Error", mock.Anything).Maybe()
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	cfg := config.OAuthConfig{
		JWTSecret:          "test-secret-key-for-testing-only",
		AccessTokenExpiry:  3600,
		RefreshTokenExpiry: 604800,
	}

	service := NewService(clientRepo, tokenRepo, cfg, logger)
	return service, clientRepo, tokenRepo
}

// Test IssueToken - Client Credentials Grant
func TestIssueToken_ClientCredentials(t *testing.T) {
	service, clientRepo, tokenRepo := setupTestService()
	ctx := context.Background()

	// Mock client
	client := &models.OAuthClient{
		ID:           "client-id-123",
		ClientID:     "test_client",
		ClientSecret: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // bcrypt hash of "secret"
		Name:         "Test Client",
		Scopes:       "read write",
		Active:       true,
	}

	// Setup mocks
	clientRepo.On("GetByClientID", ctx, "test_client").Return(client, nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.OAuthToken")).Return(nil)

	// Test
	result, err := service.IssueToken(ctx, "test_client", "secret", "client_credentials", "", "read write")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, int64(3600), result.ExpiresIn)
	assert.Equal(t, "read write", result.Scope)

	clientRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Test IssueToken - Invalid Client Credentials
func TestIssueToken_InvalidCredentials(t *testing.T) {
	service, clientRepo, _ := setupTestService()
	ctx := context.Background()

	// Mock client
	client := &models.OAuthClient{
		ID:           "client-id-123",
		ClientID:     "test_client",
		ClientSecret: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // bcrypt hash of "secret"
		Active:       true,
	}

	// Setup mocks
	clientRepo.On("GetByClientID", ctx, "test_client").Return(client, nil)

	// Test with wrong password
	result, err := service.IssueToken(ctx, "test_client", "wrong_password", "client_credentials", "", "read write")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid client credentials")

	clientRepo.AssertExpectations(t)
}

// Test IssueToken - Inactive Client
func TestIssueToken_InactiveClient(t *testing.T) {
	service, clientRepo, _ := setupTestService()
	ctx := context.Background()

	// Mock inactive client
	client := &models.OAuthClient{
		ID:           "client-id-123",
		ClientID:     "test_client",
		ClientSecret: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Active:       false,
	}

	// Setup mocks
	clientRepo.On("GetByClientID", ctx, "test_client").Return(client, nil)

	// Test
	result, err := service.IssueToken(ctx, "test_client", "secret", "client_credentials", "", "read write")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "client is not active")

	clientRepo.AssertExpectations(t)
}

// Test ValidateToken - Valid Token
func TestValidateToken_Valid(t *testing.T) {
	service, _, tokenRepo := setupTestService()
	ctx := context.Background()

	// Create a valid token
	token := &models.OAuthToken{
		ID:           "token-id-123",
		ClientID:     "client-id-123",
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		Scope:        "read write",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	// Setup mocks
	tokenRepo.On("GetByAccessToken", ctx, "valid-access-token").Return(token, nil)

	// Test
	claims, err := service.ValidateToken(ctx, "valid-access-token")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "client-id-123", claims.ClientID)
	assert.Equal(t, "read write", claims.Scope)

	tokenRepo.AssertExpectations(t)
}

// Test ValidateToken - Expired Token
func TestValidateToken_Expired(t *testing.T) {
	service, _, tokenRepo := setupTestService()
	ctx := context.Background()

	// Create an expired token
	token := &models.OAuthToken{
		ID:           "token-id-123",
		ClientID:     "client-id-123",
		AccessToken:  "expired-access-token",
		RefreshToken: "expired-refresh-token",
		Scope:        "read write",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}

	// Setup mocks
	tokenRepo.On("GetByAccessToken", ctx, "expired-access-token").Return(token, nil)

	// Test
	claims, err := service.ValidateToken(ctx, "expired-access-token")

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "token has expired")

	tokenRepo.AssertExpectations(t)
}

// Test RevokeToken - Success
func TestRevokeToken_Success(t *testing.T) {
	service, _, tokenRepo := setupTestService()
	ctx := context.Background()

	// Setup mocks
	tokenRepo.On("RevokeByAccessToken", ctx, "token-to-revoke").Return(nil)

	// Test
	err := service.RevokeToken(ctx, "token-to-revoke")

	// Assertions
	assert.NoError(t, err)

	tokenRepo.AssertExpectations(t)
}

// Test RefreshToken - Success
func TestRefreshToken_Success(t *testing.T) {
	service, clientRepo, tokenRepo := setupTestService()
	ctx := context.Background()

	// Mock client
	client := &models.OAuthClient{
		ID:           "client-id-123",
		ClientID:     "test_client",
		ClientSecret: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Name:         "Test Client",
		Scopes:       "read write",
		Active:       true,
	}

	// Mock existing token
	oldToken := &models.OAuthToken{
		ID:           "token-id-123",
		ClientID:     "client-id-123",
		AccessToken:  "old-access-token",
		RefreshToken: "valid-refresh-token",
		Scope:        "read write",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}

	// Setup mocks
	tokenRepo.On("GetByRefreshToken", ctx, "valid-refresh-token").Return(oldToken, nil)
	clientRepo.On("GetByID", ctx, "client-id-123").Return(client, nil)
	tokenRepo.On("RevokeByRefreshToken", ctx, "valid-refresh-token").Return(nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.OAuthToken")).Return(nil)

	// Test
	result, err := service.RefreshToken(ctx, "valid-refresh-token")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEqual(t, "old-access-token", result.AccessToken)
	assert.NotEqual(t, "valid-refresh-token", result.RefreshToken)

	clientRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Benchmark token generation
func BenchmarkIssueToken(b *testing.B) {
	service, clientRepo, tokenRepo := setupTestService()
	ctx := context.Background()

	client := &models.OAuthClient{
		ID:           "client-id-123",
		ClientID:     "test_client",
		ClientSecret: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Active:       true,
	}

	clientRepo.On("GetByClientID", ctx, "test_client").Return(client, nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.OAuthToken")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.IssueToken(ctx, "test_client", "secret", "client_credentials", "", "read write")
	}
}
