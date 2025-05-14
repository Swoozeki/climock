package mock_test

import (
	"testing"

	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/logger"
	"github.com/mockoho/mockoho/internal/mock"
)

func init() {
	// Initialize test logger to prevent nil pointer dereferences
	logger.InitTestLogger()
}

// createTestConfig creates a test configuration for mock tests
func createTestConfig() *config.Config {
	cfg := config.New("")

	// Set up a test feature with endpoints
	feature := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "simple-endpoint",
				Method:          "GET",
				Path:            "/api/simple",
				Active:          true,
				DefaultResponse: "standard",
				Responses: map[string]config.Response{
					"standard": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"message": "Hello, world!",
						},
						Delay: 0,
					},
					"error": {
						Status: 500,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"error": "Internal Server Error",
						},
						Delay: 0,
					},
				},
			},
			{
				ID:              "param-endpoint",
				Method:          "GET",
				Path:            "/api/users/:id",
				Active:          true,
				DefaultResponse: "standard",
				Responses: map[string]config.Response{
					"standard": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]interface{}{
							"id":   "{{params.id}}",
							"name": "User {{params.id}}",
							"date": "{{now}}",
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
				DefaultResponse: "standard",
				Responses: map[string]config.Response{
					"standard": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"message": "This endpoint is inactive",
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

// TestFindEndpoint tests the FindEndpoint function
func TestFindEndpoint(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	tests := []struct {
		name           string
		method         string
		path           string
		expectEndpoint bool
		expectedID     string
		expectedFeature string
	}{
		{
			name:           "Simple endpoint match",
			method:         "GET",
			path:           "/api/simple",
			expectEndpoint: true,
			expectedID:     "simple-endpoint",
			expectedFeature: "test",
		},
		{
			name:           "Path parameter endpoint match",
			method:         "GET",
			path:           "/api/users/123",
			expectEndpoint: true,
			expectedID:     "param-endpoint",
			expectedFeature: "test",
		},
		{
			name:           "Inactive endpoint match",
			method:         "GET",
			path:           "/api/inactive",
			expectEndpoint: true,
			expectedID:     "inactive-endpoint",
			expectedFeature: "test",
		},
		{
			name:           "Method mismatch",
			method:         "POST",
			path:           "/api/simple",
			expectEndpoint: false,
		},
		{
			name:           "Path mismatch",
			method:         "GET",
			path:           "/api/nonexistent",
			expectEndpoint: false,
		},
		{
			name:           "Path segment count mismatch",
			method:         "GET",
			path:           "/api/users/123/details",
			expectEndpoint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, feature, err := manager.FindEndpoint(tt.method, tt.path)
			
			if tt.expectEndpoint {
				if err != nil {
					t.Fatalf("Expected to find endpoint, got error: %v", err)
				}
				if endpoint == nil {
					t.Fatal("Expected non-nil endpoint")
				}
				if endpoint.ID != tt.expectedID {
					t.Errorf("Expected endpoint ID %q, got %q", tt.expectedID, endpoint.ID)
				}
				if feature != tt.expectedFeature {
					t.Errorf("Expected feature %q, got %q", tt.expectedFeature, feature)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if endpoint != nil {
					t.Errorf("Expected nil endpoint, got %v", endpoint)
				}
			}
		})
	}
}

// TestPathMatching tests path matching through FindEndpoint
// We can't test pathMatches directly as it's unexported
func TestPathMatching(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	tests := []struct {
		name     string
		method   string
		path     string
		shouldMatch bool
		expectedID string
	}{
		{
			name:     "Exact match",
			method:   "GET",
			path:     "/api/simple",
			shouldMatch: true,
			expectedID: "simple-endpoint",
		},
		{
			name:     "Parameter match",
			method:   "GET",
			path:     "/api/users/123",
			shouldMatch: true,
			expectedID: "param-endpoint",
		},
		{
			name:     "Path segment count mismatch",
			method:   "GET",
			path:     "/api/users/123/details",
			shouldMatch: false,
		},
		{
			name:     "Path mismatch",
			method:   "GET",
			path:     "/api/products/123",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, _, err := manager.FindEndpoint(tt.method, tt.path)
			
			if tt.shouldMatch {
				if err != nil {
					t.Errorf("Expected to find endpoint for %s %s, got error: %v", tt.method, tt.path, err)
				}
				if endpoint == nil {
					t.Errorf("Expected non-nil endpoint for %s %s", tt.method, tt.path)
				} else if endpoint.ID != tt.expectedID {
					t.Errorf("Expected endpoint ID %q, got %q", tt.expectedID, endpoint.ID)
				}
			} else {
				if err == nil {
					t.Errorf("Expected no match for %s %s, but got endpoint: %v", tt.method, tt.path, endpoint)
				}
			}
		})
	}
}

// TestExtractParams tests the ExtractParams function
func TestExtractParams(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	tests := []struct {
		name     string
		pattern  string
		path     string
		expected map[string]string
	}{
		{
			name:     "Single parameter",
			pattern:  "/api/users/:id",
			path:     "/api/users/123",
			expected: map[string]string{"id": "123"},
		},
		{
			name:     "Multiple parameters",
			pattern:  "/api/:resource/:id",
			path:     "/api/users/123",
			expected: map[string]string{"resource": "users", "id": "123"},
		},
		{
			name:     "No parameters",
			pattern:  "/api/simple",
			path:     "/api/simple",
			expected: map[string]string{},
		},
		{
			name:     "Mixed static and parameter segments",
			pattern:  "/api/:version/users/:id/profile",
			path:     "/api/v1/users/123/profile",
			expected: map[string]string{"version": "v1", "id": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := manager.ExtractParams(tt.pattern, tt.path)
			
			if len(params) != len(tt.expected) {
				t.Errorf("Expected %d parameters, got %d", len(tt.expected), len(params))
			}
			
			for key, expectedValue := range tt.expected {
				if value, ok := params[key]; !ok {
					t.Errorf("Expected parameter %q not found", key)
				} else if value != expectedValue {
					t.Errorf("Expected parameter %q to be %q, got %q", key, expectedValue, value)
				}
			}
		})
	}
}

// TestGenerateResponse tests the GenerateResponse function
func TestGenerateResponse(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Get the endpoint for testing
	endpoint, err := cfg.GetEndpoint("test", "simple-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}

	// Test with parameters
	params := map[string]string{"id": "123"}
	response, err := manager.GenerateResponse(endpoint, params)
	if err != nil {
		t.Fatalf("Failed to generate response: %v", err)
	}

	// Check response status and headers
	if response.Status != 200 {
		t.Errorf("Expected status 200, got %d", response.Status)
	}
	if contentType, ok := response.Headers["Content-Type"]; !ok || contentType != "application/json" {
		t.Errorf("Expected Content-Type header to be 'application/json', got %q", contentType)
	}

	// Check response body
	body, ok := response.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected body to be a map[string]interface{}, got %T", response.Body)
	}

	// Check message
	if message, ok := body["message"].(string); !ok || message != "Hello, world!" {
		t.Errorf("Expected message to be 'Hello, world!', got %v", body["message"])
	}

	// Test with non-existent response name
	endpoint.DefaultResponse = "non-existent"
	_, err = manager.GenerateResponse(endpoint, params)
	if err == nil {
		t.Error("Expected error for non-existent response, got nil")
	}
}

// TestToggleEndpoint tests the ToggleEndpoint function
func TestToggleEndpoint(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Get initial state
	endpoint, err := cfg.GetEndpoint("test", "simple-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}
	initialActive := endpoint.Active

	// Toggle endpoint
	if err := manager.ToggleEndpoint("test", "simple-endpoint"); err != nil {
		t.Fatalf("Failed to toggle endpoint: %v", err)
	}

	// Check that the state was toggled
	endpoint, err = cfg.GetEndpoint("test", "simple-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint after toggle: %v", err)
	}
	if endpoint.Active == initialActive {
		t.Errorf("Expected active state to be toggled from %v", initialActive)
	}

	// Toggle back
	if err := manager.ToggleEndpoint("test", "simple-endpoint"); err != nil {
		t.Fatalf("Failed to toggle endpoint back: %v", err)
	}

	// Check that the state was toggled back
	endpoint, err = cfg.GetEndpoint("test", "simple-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint after second toggle: %v", err)
	}
	if endpoint.Active != initialActive {
		t.Errorf("Expected active state to be toggled back to %v", initialActive)
	}

	// Test with non-existent endpoint
	if err := manager.ToggleEndpoint("test", "non-existent"); err == nil {
		t.Error("Expected error for non-existent endpoint, got nil")
	}
}

// TestSetDefaultResponse tests the SetDefaultResponse function
func TestSetDefaultResponse(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Test setting a valid response
	if err := manager.SetDefaultResponse("test", "simple-endpoint", "error"); err != nil {
		t.Fatalf("Failed to set default response: %v", err)
	}

	// Check that the default response was updated
	endpoint, err := cfg.GetEndpoint("test", "simple-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}
	if endpoint.DefaultResponse != "error" {
		t.Errorf("Expected default response to be 'error', got %q", endpoint.DefaultResponse)
	}

	// Test with non-existent response
	if err := manager.SetDefaultResponse("test", "simple-endpoint", "non-existent"); err == nil {
		t.Error("Expected error for non-existent response, got nil")
	}

	// Test with non-existent endpoint
	if err := manager.SetDefaultResponse("test", "non-existent", "standard"); err == nil {
		t.Error("Expected error for non-existent endpoint, got nil")
	}
}

// TestCreateEndpoint tests the CreateEndpoint function
func TestCreateEndpoint(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Create a new endpoint
	newEndpoint := config.Endpoint{
		ID:              "new-endpoint",
		Method:          "POST",
		Path:            "/api/new",
		Active:          true,
		DefaultResponse: "standard",
		Responses: map[string]config.Response{
			"standard": {
				Status: 201,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]string{
					"message": "Created",
				},
				Delay: 0,
			},
		},
	}

	if err := manager.CreateEndpoint("test", newEndpoint); err != nil {
		t.Fatalf("Failed to create endpoint: %v", err)
	}

	// Check that the endpoint was created
	endpoint, err := cfg.GetEndpoint("test", "new-endpoint")
	if err != nil {
		t.Fatalf("Failed to get created endpoint: %v", err)
	}
	if endpoint.ID != "new-endpoint" {
		t.Errorf("Expected endpoint ID to be 'new-endpoint', got %q", endpoint.ID)
	}
	if endpoint.Method != "POST" {
		t.Errorf("Expected endpoint method to be 'POST', got %q", endpoint.Method)
	}

	// Test creating an endpoint with duplicate ID
	if err := manager.CreateEndpoint("test", newEndpoint); err == nil {
		t.Error("Expected error for duplicate endpoint ID, got nil")
	}
}

// TestCreateFeature tests the CreateFeature function
func TestCreateFeature(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Create a new feature
	newFeature := config.FeatureConfig{
		Feature:   "new-feature",
		Endpoints: []config.Endpoint{},
	}

	if err := manager.CreateFeature(newFeature); err != nil {
		t.Fatalf("Failed to create feature: %v", err)
	}

	// Check that the feature was created
	if _, ok := cfg.Mocks["new-feature"]; !ok {
		t.Fatal("Expected feature 'new-feature' to be created")
	}

	// Test creating a feature with duplicate name
	if err := manager.CreateFeature(newFeature); err == nil {
		t.Error("Expected error for duplicate feature name, got nil")
	}
}

// TestDeleteEndpoint tests the DeleteEndpoint function
func TestDeleteEndpoint(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Delete an endpoint
	if err := manager.DeleteEndpoint("test", "simple-endpoint"); err != nil {
		t.Fatalf("Failed to delete endpoint: %v", err)
	}

	// Check that the endpoint was deleted
	_, err := cfg.GetEndpoint("test", "simple-endpoint")
	if err == nil {
		t.Fatal("Expected error for deleted endpoint, got nil")
	}

	// Test deleting a non-existent endpoint
	if err := manager.DeleteEndpoint("test", "non-existent"); err == nil {
		t.Error("Expected error for non-existent endpoint, got nil")
	}
}

// TestDeleteFeature tests the DeleteFeature function
func TestDeleteFeature(t *testing.T) {
	cfg := createTestConfig()
	manager := mock.New(cfg)

	// Delete a feature
	if err := manager.DeleteFeature("test"); err != nil {
		t.Fatalf("Failed to delete feature: %v", err)
	}

	// Check that the feature was deleted
	if _, ok := cfg.Mocks["test"]; ok {
		t.Fatal("Expected feature 'test' to be deleted")
	}

	// Test deleting a non-existent feature
	if err := manager.DeleteFeature("non-existent"); err == nil {
		t.Error("Expected error for non-existent feature, got nil")
	}
}