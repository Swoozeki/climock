package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/logger"
	"github.com/mockoho/mockoho/internal/mock"
	"github.com/mockoho/mockoho/internal/proxy"
	"github.com/mockoho/mockoho/internal/server"
)

func init() {
	// Initialize test logger to prevent nil pointer dereferences
	logger.InitTestLogger()
}

// setupTestServer creates a test server with mock endpoints
func setupTestServer(t *testing.T) (*server.Server, *httptest.Server) {
	// Create a test config with mock endpoints
	cfg := createTestConfig()

	// Create a mock manager
	mockManager := mock.New(cfg)

	// Create a real server to proxy to
	realServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"source": "real-server",
			"path":   r.URL.Path,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}))

	// Update config to use the real server as proxy target
	cfg.Global.ProxyConfig.Target = realServer.URL

	// Create a proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Create the server
	srv := server.New(cfg, mockManager, proxyManager)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	return srv, realServer
}

// createTestConfig creates a test configuration with mock endpoints
func createTestConfig() *config.Config {
	cfg := config.New("") // In-memory config

	// Set up global config
	cfg.Global = config.GlobalConfig{
		ServerConfig: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		ProxyConfig: config.ProxyConfig{
			Target:       "http://localhost:9000", // Will be overridden
			ChangeOrigin: true,
			PathRewrite:  map[string]string{},
		},
	}

	// Create a feature with endpoints
	feature := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "active-endpoint",
				Method:          "GET",
				Path:            "/api/active",
				Active:          true,
				DefaultResponse: "success",
				Responses: map[string]config.Response{
					"success": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"source": "mock-server",
							"status": "success",
						},
						Delay: 0,
					},
				},
			},
			{
				ID:              "inactive-endpoint",
				Method:          "GET",
				Path:            "/api/inactive",
				Active:          false,
				DefaultResponse: "success",
				Responses: map[string]config.Response{
					"success": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"source": "mock-server",
							"status": "success",
						},
						Delay: 0,
					},
				},
			},
		},
	}

	cfg.Mocks = map[string]config.FeatureConfig{
		"test": feature,
	}

	return cfg
}

// TestActiveEndpoint tests that the server returns mocked responses when an endpoint is active
func TestActiveEndpoint(t *testing.T) {
	// Set up test server
	srv, realServer := setupTestServer(t)
	defer func() {
		if err := srv.Stop(); err != nil {
			t.Logf("Error stopping server: %v", err)
		}
	}()
	defer realServer.Close()

	// Create a test request to the active endpoint
	req, err := http.NewRequest("GET", "http://"+srv.GetAddress()+"/api/active", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse the response body
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check that the response came from the mock server
	if source, ok := body["source"]; !ok || source != "mock-server" {
		t.Errorf("Expected source to be 'mock-server', got %q", source)
	}
}

// TestInactiveEndpoint tests that the server acts as a proxy when an endpoint is inactive
func TestInactiveEndpoint(t *testing.T) {
	// Set up test server
	srv, realServer := setupTestServer(t)
	defer func() {
		if err := srv.Stop(); err != nil {
			t.Logf("Error stopping server: %v", err)
		}
	}()
	defer realServer.Close()

	// Create a test request to the inactive endpoint
	req, err := http.NewRequest("GET", "http://"+srv.GetAddress()+"/api/inactive", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse the response body
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check that the response came from the real server
	if source, ok := body["source"]; !ok || source != "real-server" {
		t.Errorf("Expected source to be 'real-server', got %q", source)
	}
}

// TestNonConfiguredEndpoint tests that the server acts as a proxy when an endpoint is not configured
func TestNonConfiguredEndpoint(t *testing.T) {
	// Set up test server
	srv, realServer := setupTestServer(t)
	defer func() {
		if err := srv.Stop(); err != nil {
			t.Logf("Error stopping server: %v", err)
		}
	}()
	defer realServer.Close()

	// Create a test request to a non-configured endpoint
	req, err := http.NewRequest("GET", "http://"+srv.GetAddress()+"/api/non-configured", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse the response body
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check that the response came from the real server
	if source, ok := body["source"]; !ok || source != "real-server" {
		t.Errorf("Expected source to be 'real-server', got %q", source)
	}

	// Check that the path is correct
	if path, ok := body["path"]; !ok || path != "/api/non-configured" {
		t.Errorf("Expected path to be '/api/non-configured', got %q", path)
	}
}

// TestPathParameters tests that the server correctly handles path parameters
func TestPathParameters(t *testing.T) {
	// Create a test config with path parameter endpoint
	cfg := config.New("") // In-memory config
	cfg.Global = config.GlobalConfig{
		ServerConfig: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		ProxyConfig: config.ProxyConfig{
			Target:       "http://localhost:9000",
			ChangeOrigin: true,
			PathRewrite:  map[string]string{},
		},
	}

	// Create a feature with an endpoint that has path parameters
	feature := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "user-endpoint",
				Method:          "GET",
				Path:            "/api/users/:id",
				Active:          true,
				DefaultResponse: "success",
				Responses: map[string]config.Response{
					"success": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]interface{}{
							"source": "mock-server",
							"userId": "{{.params.id}}",
						},
						Delay: 0,
					},
				},
			},
		},
	}

	cfg.Mocks = map[string]config.FeatureConfig{
		"test": feature,
	}

	// Create a mock manager
	mockManager := mock.New(cfg)

	// Create a real server to proxy to
	realServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"source": "real-server",
			"path":   r.URL.Path,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}))
	defer realServer.Close()

	// Update config to use the real server as proxy target
	cfg.Global.ProxyConfig.Target = realServer.URL

	// Create a proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Create the server
	srv := server.New(cfg, mockManager, proxyManager)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Create a test request to the endpoint with a path parameter
	req, err := http.NewRequest("GET", "http://"+srv.GetAddress()+"/api/users/123", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse the response body
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	// Check that the response came from the mock server
	if source, ok := body["source"]; !ok || source != "mock-server" {
		t.Errorf("Expected source to be 'mock-server', got %q", source)
	}

	// Check that the path parameter was correctly extracted
	if userId, ok := body["userId"]; !ok || userId != "123" {
		t.Errorf("Expected userId to be '123', got %q", userId)
	}
}

// TestDifferentMethods tests that the server correctly handles different HTTP methods
func TestDifferentMethods(t *testing.T) {
	// Create a test config with different HTTP methods
	cfg := config.New("") // In-memory config
	cfg.Global = config.GlobalConfig{
		ServerConfig: config.ServerConfig{
			Port: 8080,
			Host: "localhost",
		},
		ProxyConfig: config.ProxyConfig{
			Target:       "http://localhost:9000",
			ChangeOrigin: true,
			PathRewrite:  map[string]string{},
		},
	}

	// Create a feature with endpoints for different HTTP methods
	feature := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "post-endpoint",
				Method:          "POST",
				Path:            "/api/resource",
				Active:          true,
				DefaultResponse: "created",
				Responses: map[string]config.Response{
					"created": {
						Status: 201,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"source": "mock-server",
							"method": "POST",
							"status": "created",
						},
						Delay: 0,
					},
				},
			},
			{
				ID:              "put-endpoint",
				Method:          "PUT",
				Path:            "/api/resource",
				Active:          true,
				DefaultResponse: "updated",
				Responses: map[string]config.Response{
					"updated": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"source": "mock-server",
							"method": "PUT",
							"status": "updated",
						},
						Delay: 0,
					},
				},
			},
			{
				ID:              "delete-endpoint",
				Method:          "DELETE",
				Path:            "/api/resource",
				Active:          true,
				DefaultResponse: "deleted",
				Responses: map[string]config.Response{
					"deleted": {
						Status: 204,
						Headers: map[string]string{},
						Body:    nil,
						Delay:   0,
					},
				},
			},
		},
	}

	cfg.Mocks = map[string]config.FeatureConfig{
		"test": feature,
	}

	// Create a mock manager
	mockManager := mock.New(cfg)

	// Create a real server to proxy to
	realServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"source": "real-server",
			"method": r.Method,
		})
	}))
	defer realServer.Close()

	// Update config to use the real server as proxy target
	cfg.Global.ProxyConfig.Target = realServer.URL

	// Create a proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Create the server
	srv := server.New(cfg, mockManager, proxyManager)

	// Start the server
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Test POST request
	t.Run("POST", func(t *testing.T) {
		req, err := http.NewRequest("POST", "http://"+srv.GetAddress()+"/api/resource", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to parse response body: %v", err)
		}

		if source, ok := body["source"]; !ok || source != "mock-server" {
			t.Errorf("Expected source to be 'mock-server', got %q", source)
		}

		if method, ok := body["method"]; !ok || method != "POST" {
			t.Errorf("Expected method to be 'POST', got %q", method)
		}
	})

	// Test PUT request
	t.Run("PUT", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "http://"+srv.GetAddress()+"/api/resource", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to parse response body: %v", err)
		}

		if source, ok := body["source"]; !ok || source != "mock-server" {
			t.Errorf("Expected source to be 'mock-server', got %q", source)
		}

		if method, ok := body["method"]; !ok || method != "PUT" {
			t.Errorf("Expected method to be 'PUT', got %q", method)
		}
	})

	// Test DELETE request
	t.Run("DELETE", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", "http://"+srv.GetAddress()+"/api/resource", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status code %d, got %d", http.StatusNoContent, resp.StatusCode)
		}
	})
}