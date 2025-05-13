package tests

import (
	"testing"

	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/mock"
)

func TestFindEndpoint(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Mocks: map[string]config.FeatureConfig{
			"users": {
				Feature: "users",
				Endpoints: []config.Endpoint{
					{
						ID:              "get-user",
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
									"id":   "123",
									"name": "Test User",
								},
							},
						},
					},
				},
			},
		},
	}

	// Create mock manager
	mockManager := mock.New(cfg)

	// Test finding an endpoint
	endpoint, feature, err := mockManager.FindEndpoint("GET", "/api/users/123")
	if err != nil {
		t.Errorf("Failed to find endpoint: %v", err)
	}
	if feature != "users" {
		t.Errorf("Expected feature 'users', got '%s'", feature)
	}
	if endpoint.ID != "get-user" {
		t.Errorf("Expected endpoint 'get-user', got '%s'", endpoint.ID)
	}

	// Test finding a non-existent endpoint
	_, _, err = mockManager.FindEndpoint("POST", "/api/users/123")
	if err == nil {
		t.Error("Expected error for non-existent endpoint, got nil")
	}
}

func TestExtractParams(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{}

	// Create mock manager
	mockManager := mock.New(cfg)

	// Test extracting parameters
	params := mockManager.ExtractParams("/api/users/:id", "/api/users/123")
	if params["id"] != "123" {
		t.Errorf("Expected id '123', got '%s'", params["id"])
	}

	// Test extracting multiple parameters
	params = mockManager.ExtractParams("/api/users/:id/posts/:postId", "/api/users/123/posts/456")
	if params["id"] != "123" {
		t.Errorf("Expected id '123', got '%s'", params["id"])
	}
	if params["postId"] != "456" {
		t.Errorf("Expected postId '456', got '%s'", params["postId"])
	}
}

func TestGenerateResponse(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{}

	// Create mock manager
	mockManager := mock.New(cfg)

	// Create a test endpoint
	endpoint := &config.Endpoint{
		ID:              "test-endpoint",
		Method:          "GET",
		Path:            "/api/test/:id",
		Active:          true,
		DefaultResponse: "standard",
		Responses: map[string]config.Response{
			"standard": {
				Status: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"id":      "{{params.id}}",
					"message": "Hello, World!",
				},
				Delay: 0,
			},
		},
	}

	// Test generating a response
	params := map[string]string{
		"id": "123",
	}
	response, err := mockManager.GenerateResponse(endpoint, params)
	if err != nil {
		t.Errorf("Failed to generate response: %v", err)
	}
	if response.Status != 200 {
		t.Errorf("Expected status 200, got %d", response.Status)
	}
	if response.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", response.Headers["Content-Type"])
	}

	// Check that the template variable was replaced
	body, ok := response.Body.(map[string]interface{})
	if !ok {
		t.Error("Expected body to be a map")
		return
	}
	if body["id"] != "123" {
		t.Errorf("Expected id '123', got '%v'", body["id"])
	}
}