// Package oauth provides OAuth 2.0 service implementation.
package oauth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// Service handles OAuth 2.0 business logic
type Service struct {
	clientRepo *repositories.OAuthClientRepository
	tokenRepo  *repositories.OAuthTokenRepository
	tokenGen   *TokenGenerator
	logger     *logger.Logger
}

// NewService creates a new OAuth service with JWT secret validation
func NewService(
	clientRepo *repositories.OAuthClientRepository,
	tokenRepo *repositories.OAuthTokenRepository,
	cfg config.OAuthConfig,
	log *logger.Logger,
) (*Service, error) {
	// Create token generator with secret validation
	tokenGen, err := NewTokenGenerator(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create token generator: %w", err)
	}

	return &Service{
		clientRepo: clientRepo,
		tokenRepo:  tokenRepo,
		tokenGen:   tokenGen,
		logger:     log,
	}, nil
}

// TokenRequest represents a token request
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// HandleTokenRequest handles OAuth token requests
func (s *Service) HandleTokenRequest(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	switch req.GrantType {
	case "client_credentials":
		return s.handleClientCredentials(ctx, req)
	case "refresh_token":
		return s.handleRefreshToken(ctx, req)
	case "authorization_code":
		return s.handleAuthorizationCode(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported grant type: %s", req.GrantType)
	}
}

// handleClientCredentials handles client credentials grant
func (s *Service) handleClientCredentials(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	// Validate client credentials
	client, err := s.clientRepo.GetByClientID(ctx, req.ClientID)
	if err != nil {
		s.logger.Warnf("Client not found: %s", req.ClientID)
		return nil, fmt.Errorf("invalid client credentials")
	}

	if !client.Enabled {
		s.logger.Warnf("Client disabled: %s", req.ClientID)
		return nil, fmt.Errorf("client is disabled")
	}

	// Validate client secret
	if !s.clientRepo.ValidateSecret(client, req.ClientSecret) {
		s.logger.Warnf("Invalid client secret for: %s", req.ClientID)
		return nil, fmt.Errorf("invalid client credentials")
	}

	// Use provided scope or default scope
	scope := req.Scope
	if scope == "" {
		scope = client.DefaultScope
	}

	// Generate access token
	accessToken, expiresAt, err := s.tokenGen.GenerateAccessToken(client.ClientID, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err := s.tokenGen.GenerateRefreshToken(client.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store token in database
	token := &repositories.OAuthToken{
		ID:           uuid.New(),
		ClientID:     client.ID,
		AccessToken:  accessToken,
		RefreshToken: &refreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    &expiresAt,
		Scope:        scope,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		s.logger.Error(err, "Failed to store token")
		// Continue anyway - token is still valid
	}

	s.logger.Infof("Issued access token for client: %s", client.ClientID)

	expiresIn := time.Until(expiresAt).Seconds()
	if expiresIn < 0 {
		expiresIn = 0
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(expiresIn),
		RefreshToken: refreshToken,
		Scope:        scope,
	}, nil
}

// handleRefreshToken handles refresh token grant
func (s *Service) handleRefreshToken(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := s.tokenGen.ValidateToken(req.RefreshToken)
	if err != nil {
		s.logger.Warnf("Invalid refresh token: %v", err)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get client
	client, err := s.clientRepo.GetByClientID(ctx, claims.ClientID)
	if err != nil {
		return nil, fmt.Errorf("client not found")
	}

	if !client.Enabled {
		return nil, fmt.Errorf("client is disabled")
	}

	// Generate new access token
	scope := req.Scope
	if scope == "" {
		scope = claims.Scope
		if scope == "refresh" {
			scope = client.DefaultScope
		}
	}

	accessToken, expiresAt, err := s.tokenGen.GenerateAccessToken(client.ClientID, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, _, err := s.tokenGen.GenerateRefreshToken(client.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store new token
	token := &repositories.OAuthToken{
		ID:           uuid.New(),
		ClientID:     client.ID,
		AccessToken:  accessToken,
		RefreshToken: &newRefreshToken,
		TokenType:    "Bearer",
		ExpiresAt:    &expiresAt,
		Scope:        scope,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		s.logger.Error(err, "Failed to store token")
	}

	// Revoke old refresh token
	if err := s.tokenRepo.DeleteByRefreshToken(ctx, req.RefreshToken); err != nil {
		s.logger.Error(err, "Failed to revoke old refresh token")
	}

	s.logger.Infof("Refreshed access token for client: %s", client.ClientID)

	expiresIn := time.Until(expiresAt).Seconds()
	if expiresIn < 0 {
		expiresIn = 0
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(expiresIn),
		RefreshToken: newRefreshToken,
		Scope:        scope,
	}, nil
}

// handleAuthorizationCode handles authorization code grant (placeholder)
func (s *Service) handleAuthorizationCode(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	// TODO: Implement authorization code flow
	// For now, return error
	return nil, fmt.Errorf("authorization code grant not yet implemented")
}

// ValidateAccessToken validates an access token and returns the claims
func (s *Service) ValidateAccessToken(ctx context.Context, accessToken string) (*Claims, error) {
	// Validate JWT signature and claims
	claims, err := s.tokenGen.ValidateToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if client still exists and is enabled
	client, err := s.clientRepo.GetByClientID(ctx, claims.ClientID)
	if err != nil {
		return nil, fmt.Errorf("client not found")
	}

	if !client.Enabled {
		return nil, fmt.Errorf("client is disabled")
	}

	return claims, nil
}

// RevokeToken revokes an access token
func (s *Service) RevokeToken(ctx context.Context, token string) error {
	// Try to revoke as access token
	if err := s.tokenRepo.DeleteByAccessToken(ctx, token); err == nil {
		s.logger.Infof("Revoked access token")
		return nil
	}

	// Try to revoke as refresh token
	if err := s.tokenRepo.DeleteByRefreshToken(ctx, token); err == nil {
		s.logger.Infof("Revoked refresh token")
		return nil
	}

	return fmt.Errorf("token not found")
}

// CleanupExpiredTokens removes expired tokens from the database
func (s *Service) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	count, err := s.tokenRepo.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	if count > 0 {
		s.logger.Infof("Cleaned up %d expired tokens", count)
	}

	return count, nil
}
