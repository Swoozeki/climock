package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mockoho/mockoho/internal/config"
)

func TestConfigLoad(t *testing.T) {
	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "mockoho-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a global config file
	globalConfig := config.GlobalConfig{
		ProxyConfig: config.ProxyConfig{
			Target:       "https://api.example.com",
			ChangeOrigin: true,
			PathRewrite: map[string]string{
				"^/api": "",
			},
		},
		ServerConfig: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
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
	err = os.WriteFile(filepath.Join(tempDir, "config.json"), globalConfigData, 0644)
	if err != nil {
		t.Fatalf("Failed to write global config: %v", err)
	}

	// Create a feature config file
	featureConfig := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "test-endpoint",
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
						Body: map[string]interface{}{
							"message": "Hello, World!",
						},
						Delay: 0,
					},
				},
			},
		},
	}
	featureConfigData, err := json.MarshalIndent(featureConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal feature config: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "test.json"), featureConfigData, 0644)
	if err != nil {
		t.Fatalf("Failed to write feature config: %v", err)
	}

	// Create a config instance
	cfg := config.New(tempDir)

	// Load the configuration
	err = cfg.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify the global config was loaded correctly
	if cfg.Global.ServerConfig.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Global.ServerConfig.Port)
	}
	if cfg.Global.ProxyConfig.Target != "https://api.example.com" {
		t.Errorf("Expected target 'https://api.example.com', got '%s'", cfg.Global.ProxyConfig.Target)
	}

	// Verify the feature config was loaded correctly
	if len(cfg.Mocks) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(cfg.Mocks))
	}
	if _, ok := cfg.Mocks["test"]; !ok {
		t.Error("Expected feature 'test' to be loaded")
	}
	if len(cfg.Mocks["test"].Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(cfg.Mocks["test"].Endpoints))
	}
	if cfg.Mocks["test"].Endpoints[0].ID != "test-endpoint" {
		t.Errorf("Expected endpoint 'test-endpoint', got '%s'", cfg.Mocks["test"].Endpoints[0].ID)
	}
}

func TestConfigSave(t *testing.T) {
	// Create a temporary directory for test configs
	tempDir, err := os.MkdirTemp("", "mockoho-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a config instance
	cfg := config.New(tempDir)

	// Set up the global config
	cfg.Global = config.GlobalConfig{
		ProxyConfig: config.ProxyConfig{
			Target:       "https://api.example.com",
			ChangeOrigin: true,
			PathRewrite: map[string]string{
				"^/api": "",
			},
		},
		ServerConfig: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
	}

	// Save the global config
	err = cfg.SaveGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to save global config: %v", err)
	}

	// Verify the global config file was created
	if _, err := os.Stat(filepath.Join(tempDir, "config.json")); os.IsNotExist(err) {
		t.Error("Global config file was not created")
	}

	// Set up a feature config
	featureConfig := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "test-endpoint",
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
						Body: map[string]interface{}{
							"message": "Hello, World!",
						},
						Delay: 0,
					},
				},
			},
		},
	}
	cfg.Mocks = map[string]config.FeatureConfig{
		"test": featureConfig,
	}

	// Save the feature config
	err = cfg.SaveFeatureConfig("test")
	if err != nil {
		t.Fatalf("Failed to save feature config: %v", err)
	}

	// Verify the feature config file was created
	if _, err := os.Stat(filepath.Join(tempDir, "test.json")); os.IsNotExist(err) {
		t.Error("Feature config file was not created")
	}
}

func TestConfigEndpointOperations(t *testing.T) {
	// Create a config instance
	cfg := config.New("")

	// Set up a feature config
	featureConfig := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "test-endpoint",
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
						Body: map[string]interface{}{
							"message": "Hello, World!",
						},
						Delay: 0,
					},
				},
			},
		},
	}
	cfg.Mocks = map[string]config.FeatureConfig{
		"test": featureConfig,
	}

	// Test getting an endpoint
	endpoint, err := cfg.GetEndpoint("test", "test-endpoint")
	if err != nil {
		t.Errorf("Failed to get endpoint: %v", err)
	}
	if endpoint.ID != "test-endpoint" {
		t.Errorf("Expected endpoint 'test-endpoint', got '%s'", endpoint.ID)
	}

	// Test updating an endpoint
	endpoint.Active = false
	err = cfg.UpdateEndpoint("test", *endpoint)
	if err != nil {
		t.Errorf("Failed to update endpoint: %v", err)
	}
	updatedEndpoint, _ := cfg.GetEndpoint("test", "test-endpoint")
	if updatedEndpoint.Active != false {
		t.Error("Endpoint was not updated")
	}

	// Test adding a new endpoint
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
				Body: map[string]interface{}{
					"message": "Created",
				},
				Delay: 0,
			},
		},
	}
	err = cfg.AddEndpoint("test", newEndpoint)
	if err != nil {
		t.Errorf("Failed to add endpoint: %v", err)
	}
	if len(cfg.Mocks["test"].Endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(cfg.Mocks["test"].Endpoints))
	}

	// Test deleting an endpoint
	err = cfg.DeleteEndpoint("test", "test-endpoint")
	if err != nil {
		t.Errorf("Failed to delete endpoint: %v", err)
	}
	if len(cfg.Mocks["test"].Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(cfg.Mocks["test"].Endpoints))
	}
	if cfg.Mocks["test"].Endpoints[0].ID != "new-endpoint" {
		t.Errorf("Expected endpoint 'new-endpoint', got '%s'", cfg.Mocks["test"].Endpoints[0].ID)
	}
}