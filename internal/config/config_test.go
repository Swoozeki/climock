package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"swoozeki/climock/internal/config"
	"swoozeki/climock/internal/logger"
)

func init() {
	// Initialize test logger to prevent nil pointer dereferences
	logger.InitTestLogger()
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	cfg := config.New("testdir")
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
	if cfg.BaseDir != "testdir" {
		t.Errorf("Expected BaseDir to be 'testdir', got %q", cfg.BaseDir)
	}
	if cfg.Mocks == nil {
		t.Error("Expected non-nil Mocks map")
	}
}

// TestLoadAndSave tests loading and saving configurations
func TestLoadAndSave(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "climock-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a global config file
	globalConfig := config.GlobalConfig{
		ServerConfig: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
		ProxyConfig: config.ProxyConfig{
			Target:       "https://api.example.com",
			ChangeOrigin: true,
			PathRewrite: map[string]string{
				"^/api": "",
			},
		},
		Editor: config.EditorConfig{
			Command: "code",
			Args:    []string{"-g", "{file}:{line}"},
		},
	}

	globalConfigData, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal global config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "config.json"), globalConfigData, 0644); err != nil {
		t.Fatalf("Failed to write global config file: %v", err)
	}

	// Create a feature config file
	featureConfig := config.FeatureConfig{
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
							"id":   "{{params.id}}",
							"name": "John Doe",
						},
						Delay: 100,
					},
				},
			},
		},
	}

	featureConfigData, err := json.MarshalIndent(featureConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal feature config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "users.json"), featureConfigData, 0644); err != nil {
		t.Fatalf("Failed to write feature config file: %v", err)
	}

	// Create a config instance and load the configuration
	cfg := config.New(tempDir)
	if err := cfg.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify global config was loaded correctly
	if cfg.Global.ServerConfig.Port != 3000 {
		t.Errorf("Expected port to be 3000, got %d", cfg.Global.ServerConfig.Port)
	}
	if cfg.Global.ProxyConfig.Target != "https://api.example.com" {
		t.Errorf("Expected target to be 'https://api.example.com', got %q", cfg.Global.ProxyConfig.Target)
	}

	// Verify feature config was loaded correctly
	if len(cfg.Mocks) != 1 {
		t.Fatalf("Expected 1 feature, got %d", len(cfg.Mocks))
	}
	if _, ok := cfg.Mocks["users"]; !ok {
		t.Fatal("Expected 'users' feature to be loaded")
	}
	if len(cfg.Mocks["users"].Endpoints) != 1 {
		t.Fatalf("Expected 1 endpoint, got %d", len(cfg.Mocks["users"].Endpoints))
	}
	if cfg.Mocks["users"].Endpoints[0].ID != "get-user" {
		t.Errorf("Expected endpoint ID to be 'get-user', got %q", cfg.Mocks["users"].Endpoints[0].ID)
	}

	// Test saving global config
	cfg.Global.ServerConfig.Port = 4000
	if err := cfg.SaveGlobalConfig(); err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	// Verify the global config was saved correctly
	var savedGlobalConfig config.GlobalConfig
	savedGlobalConfigData, err := os.ReadFile(filepath.Join(tempDir, "config.json"))
	if err != nil {
		t.Fatalf("Failed to read saved global config: %v", err)
	}
	if err := json.Unmarshal(savedGlobalConfigData, &savedGlobalConfig); err != nil {
		t.Fatalf("Failed to unmarshal saved global config: %v", err)
	}
	if savedGlobalConfig.ServerConfig.Port != 4000 {
		t.Errorf("Expected saved port to be 4000, got %d", savedGlobalConfig.ServerConfig.Port)
	}

	// Test saving feature config
	endpoint, err := cfg.GetEndpoint("users", "get-user")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}
	endpoint.Active = false
	if err := cfg.UpdateEndpoint("users", *endpoint); err != nil {
		t.Fatalf("Failed to update endpoint: %v", err)
	}
	if err := cfg.SaveFeatureConfig("users"); err != nil {
		t.Fatalf("Failed to save feature config: %v", err)
	}

	// Verify the feature config was saved correctly
	var savedFeatureConfig config.FeatureConfig
	savedFeatureConfigData, err := os.ReadFile(filepath.Join(tempDir, "users.json"))
	if err != nil {
		t.Fatalf("Failed to read saved feature config: %v", err)
	}
	if err := json.Unmarshal(savedFeatureConfigData, &savedFeatureConfig); err != nil {
		t.Fatalf("Failed to unmarshal saved feature config: %v", err)
	}
	if savedFeatureConfig.Endpoints[0].Active {
		t.Error("Expected saved endpoint to be inactive")
	}
}

// TestEndpointManagement tests endpoint management functions
func TestEndpointManagement(t *testing.T) {
	cfg := config.New("")

	// Set up a test feature
	feature := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "existing-endpoint",
				Method:          "GET",
				Path:            "/api/test",
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
				},
			},
		},
	}

	cfg.Mocks = map[string]config.FeatureConfig{
		"test": feature,
	}

	// Test GetEndpoint
	endpoint, err := cfg.GetEndpoint("test", "existing-endpoint")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
	}
	if endpoint.ID != "existing-endpoint" {
		t.Errorf("Expected endpoint ID to be 'existing-endpoint', got %q", endpoint.ID)
	}

	// Test GetEndpoint with non-existent feature
	_, err = cfg.GetEndpoint("non-existent", "existing-endpoint")
	if err == nil {
		t.Error("Expected error for non-existent feature, got nil")
	}

	// Test GetEndpoint with non-existent endpoint
	_, err = cfg.GetEndpoint("test", "non-existent")
	if err == nil {
		t.Error("Expected error for non-existent endpoint, got nil")
	}

	// Test UpdateEndpoint
	endpoint.Active = false
	if err := cfg.UpdateEndpoint("test", *endpoint); err != nil {
		t.Fatalf("Failed to update endpoint: %v", err)
	}
	updatedEndpoint, _ := cfg.GetEndpoint("test", "existing-endpoint")
	if updatedEndpoint.Active {
		t.Error("Expected updated endpoint to be inactive")
	}

	// Test AddEndpoint
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
	if err := cfg.AddEndpoint("test", newEndpoint); err != nil {
		t.Fatalf("Failed to add endpoint: %v", err)
	}
	if len(cfg.Mocks["test"].Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(cfg.Mocks["test"].Endpoints))
	}
	
	// Verify that the new endpoint is inactive by default, regardless of the provided value
	addedEndpoint, err := cfg.GetEndpoint("test", "new-endpoint")
	if err != nil {
		t.Fatalf("Failed to get newly added endpoint: %v", err)
	}
	if addedEndpoint.Active {
		t.Error("Expected newly added endpoint to be inactive by default")
	}

	// Test AddEndpoint with duplicate ID
	if err := cfg.AddEndpoint("test", newEndpoint); err == nil {
		t.Error("Expected error for duplicate endpoint ID, got nil")
	}

	// Test DeleteEndpoint
	if err := cfg.DeleteEndpoint("test", "new-endpoint"); err != nil {
		t.Fatalf("Failed to delete endpoint: %v", err)
	}
	if len(cfg.Mocks["test"].Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint after deletion, got %d", len(cfg.Mocks["test"].Endpoints))
	}

	// Test DeleteEndpoint with non-existent endpoint
	if err := cfg.DeleteEndpoint("test", "non-existent"); err == nil {
		t.Error("Expected error for deleting non-existent endpoint, got nil")
	}
}

// TestFeatureManagement tests feature management functions
func TestFeatureManagement(t *testing.T) {
	cfg := config.New("")

	// Set up initial state
	cfg.Mocks = map[string]config.FeatureConfig{
		"existing": {
			Feature:   "existing",
			Endpoints: []config.Endpoint{},
		},
	}

	// Test AddFeature
	newFeature := config.FeatureConfig{
		Feature:   "new-feature",
		Endpoints: []config.Endpoint{},
	}
	if err := cfg.AddFeature(newFeature); err != nil {
		t.Fatalf("Failed to add feature: %v", err)
	}
	if len(cfg.Mocks) != 2 {
		t.Errorf("Expected 2 features, got %d", len(cfg.Mocks))
	}
	if _, ok := cfg.Mocks["new-feature"]; !ok {
		t.Error("Expected 'new-feature' to be added")
	}

	// Test AddFeature with duplicate name
	if err := cfg.AddFeature(newFeature); err == nil {
		t.Error("Expected error for duplicate feature name, got nil")
	}

	// Test DeleteFeature
	if err := cfg.DeleteFeature("new-feature"); err != nil {
		t.Fatalf("Failed to delete feature: %v", err)
	}
	if len(cfg.Mocks) != 1 {
		t.Errorf("Expected 1 feature after deletion, got %d", len(cfg.Mocks))
	}
	if _, ok := cfg.Mocks["new-feature"]; ok {
		t.Error("Expected 'new-feature' to be deleted")
	}

	// Test DeleteFeature with non-existent feature
	if err := cfg.DeleteFeature("non-existent"); err == nil {
		t.Error("Expected error for deleting non-existent feature, got nil")
	}
}