//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	baseURL      = getEnv("API_BASE_URL", "http://localhost:8080")
	clientID     = getEnv("TEST_CLIENT_ID", "test_client")
	clientSecret = getEnv("TEST_CLIENT_SECRET", "test_secret_123456")
	adminAPIKey  = getEnv("ADMIN_API_KEY", "admin_dev_key_12345678901234567890123456789012")
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Test Health Endpoint
func TestHealth(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var healthResp map[string]interface{}
	err = json.Unmarshal(body, &healthResp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", healthResp["status"])
}

// Test OAuth Token Endpoint - Client Credentials Grant
func TestOAuth_ClientCredentials(t *testing.T) {
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "read write",
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(
		baseURL+"/oauth/token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var tokenResp map[string]interface{}
	err = json.Unmarshal(body, &tokenResp)
	require.NoError(t, err)

	assert.NotEmpty(t, tokenResp["access_token"])
	assert.NotEmpty(t, tokenResp["refresh_token"])
	assert.Equal(t, "Bearer", tokenResp["token_type"])
	assert.NotZero(t, tokenResp["expires_in"])
}

// Test OAuth Token Endpoint - Invalid Credentials
func TestOAuth_InvalidCredentials(t *testing.T) {
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": "wrong_password",
		"scope":         "read write",
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(
		baseURL+"/oauth/token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Helper function to get OAuth token
func getOAuthToken(t *testing.T) string {
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "read write",
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(
		baseURL+"/oauth/token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	return tokenResp["access_token"].(string)
}

// Test List Models Endpoint
func TestModels_List(t *testing.T) {
	token := getOAuthToken(t)

	req, err := http.NewRequest("GET", baseURL+"/v1/models", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var modelsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&modelsResp)
	require.NoError(t, err)

	assert.Equal(t, "list", modelsResp["object"])
	assert.NotEmpty(t, modelsResp["data"])
}

// Test Get Single Model Endpoint
func TestModels_GetSingle(t *testing.T) {
	token := getOAuthToken(t)

	req, err := http.NewRequest("GET", baseURL+"/v1/models/claude-3-opus-20240229", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var modelResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&modelResp)
	require.NoError(t, err)

	assert.Equal(t, "claude-3-opus-20240229", modelResp["id"])
	assert.Equal(t, "model", modelResp["object"])
}

// Test Chat Completions Endpoint - Without OAuth (Should Fail)
func TestChat_NoAuth(t *testing.T) {
	payload := map[string]interface{}{
		"model": "claude-3-opus-20240229",
		"messages": []map[string]string{
			{"role": "user", "content": "Say hello"},
		},
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	resp, err := http.Post(
		baseURL+"/v1/chat/completions",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Test Admin API - List Clients
func TestAdmin_ListClients(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/admin/clients", nil)
	require.NoError(t, err)

	req.Header.Set("X-Admin-API-Key", adminAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var clientsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&clientsResp)
	require.NoError(t, err)

	assert.Contains(t, clientsResp, "clients")
}

// Test Admin API - Get Single Client
func TestAdmin_GetClient(t *testing.T) {
	// First, list clients to get an ID
	req, err := http.NewRequest("GET", baseURL+"/admin/clients", nil)
	require.NoError(t, err)
	req.Header.Set("X-Admin-API-Key", adminAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var listResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)

	clients := listResp["clients"].([]interface{})
	if len(clients) == 0 {
		t.Skip("No clients available for testing")
	}

	clientData := clients[0].(map[string]interface{})
	clientID := clientData["id"].(string)

	// Now get the specific client
	req2, err := http.NewRequest("GET", baseURL+"/admin/clients/"+clientID, nil)
	require.NoError(t, err)
	req2.Header.Set("X-Admin-API-Key", adminAPIKey)

	resp2, err := client.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}

// Test Admin API - Cache Stats
func TestAdmin_CacheStats(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/admin/cache/stats", nil)
	require.NoError(t, err)

	req.Header.Set("X-Admin-API-Key", adminAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statsResp)
	require.NoError(t, err)

	assert.Contains(t, statsResp, "hits")
	assert.Contains(t, statsResp, "misses")
}

// Test Admin API - Provider Status
func TestAdmin_ProviderStatus(t *testing.T) {
	req, err := http.NewRequest("GET", baseURL+"/admin/providers/status", nil)
	require.NoError(t, err)

	req.Header.Set("X-Admin-API-Key", adminAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statusResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statusResp)
	require.NoError(t, err)

	assert.Contains(t, statusResp, "providers")
}

// Test Admin API - Without API Key (Should Fail)
func TestAdmin_NoAPIKey(t *testing.T) {
	resp, err := http.Get(baseURL + "/admin/clients")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Test Metrics Endpoint
func TestMetrics(t *testing.T) {
	resp, err := http.Get(baseURL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Check for some expected Prometheus metrics
	bodyStr := string(body)
	assert.Contains(t, bodyStr, "llm_proxy_http_requests_total")
	assert.Contains(t, bodyStr, "llm_proxy_http_request_duration_seconds")
}

// Test OAuth Token Refresh
func TestOAuth_RefreshToken(t *testing.T) {
	// First get initial tokens
	payload1 := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"scope":         "read write",
	}

	jsonData1, err := json.Marshal(payload1)
	require.NoError(t, err)

	resp1, err := http.Post(
		baseURL+"/oauth/token",
		"application/json",
		bytes.NewBuffer(jsonData1),
	)
	require.NoError(t, err)
	defer resp1.Body.Close()

	var tokenResp1 map[string]interface{}
	err = json.NewDecoder(resp1.Body).Decode(&tokenResp1)
	require.NoError(t, err)

	refreshToken := tokenResp1["refresh_token"].(string)

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Now use refresh token
	payload2 := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	jsonData2, err := json.Marshal(payload2)
	require.NoError(t, err)

	resp2, err := http.Post(
		baseURL+"/oauth/token",
		"application/json",
		bytes.NewBuffer(jsonData2),
	)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var tokenResp2 map[string]interface{}
	err = json.NewDecoder(resp2.Body).Decode(&tokenResp2)
	require.NoError(t, err)

	// Should get new tokens
	assert.NotEmpty(t, tokenResp2["access_token"])
	assert.NotEmpty(t, tokenResp2["refresh_token"])
	assert.NotEqual(t, tokenResp1["access_token"], tokenResp2["access_token"])
}

// Test Create OAuth Client via Admin API
func TestAdmin_CreateClient(t *testing.T) {
	payload := map[string]interface{}{
		"client_id":     fmt.Sprintf("test_client_%d", time.Now().Unix()),
		"client_secret": "test_secret_123",
		"name":          "Integration Test Client",
		"scopes":        "read write",
		"active":        true,
	}

	jsonData, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(
		"POST",
		baseURL+"/admin/clients",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-API-Key", adminAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(t, err)

	assert.NotEmpty(t, createResp["id"])
	assert.Equal(t, payload["name"], createResp["name"])
}

// Performance Test - Multiple OAuth Requests
func TestPerformance_MultipleOAuthRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	const iterations = 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_ = getOAuthToken(t)
	}

	duration := time.Since(start)
	avgDuration := duration / iterations

	t.Logf("Completed %d OAuth requests in %v (avg: %v per request)",
		iterations, duration, avgDuration)

	// Should be reasonably fast
	assert.Less(t, avgDuration, 500*time.Millisecond)
}
