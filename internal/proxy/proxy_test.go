package proxy_test

import (
	"testing"

	"kohofinancial/mockoho/internal/config"
	"kohofinancial/mockoho/internal/logger"
	"kohofinancial/mockoho/internal/proxy"
)

func init() {
	// Initialize test logger to prevent nil pointer dereferences
	logger.InitTestLogger()
}

// createTestConfig creates a test configuration for proxy tests
func createTestConfig() *config.Config {
	cfg := config.New(".")

	// Set up global config with proxy settings
	cfg.Global = config.GlobalConfig{
		ServerConfig: config.ServerConfig{
			Port: 3000,
			Host: "localhost",
		},
		ProxyConfig: config.ProxyConfig{
			Target:       "http://example.com",
			ChangeOrigin: true,
			PathRewrite: map[string]string{
				"^/api": "",
			},
		},
	}

	return cfg
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	cfg := createTestConfig()

	// Test with valid target URL
	manager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	if manager == nil {
		t.Fatal("Expected non-nil proxy manager")
	}

	// Test with invalid target URL
	cfg.Global.ProxyConfig.Target = "://invalid-url"
	_, err = proxy.New(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid target URL, got nil")
	}
}

// TestUpdateTarget tests the UpdateTarget function
func TestUpdateTarget(t *testing.T) {
	cfg := createTestConfig()
	manager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Update the target
	newTarget := "https://api.example.org"
	if err := manager.UpdateTarget(newTarget); err != nil {
		t.Fatalf("Failed to update target: %v", err)
	}

	// Check that the target was updated
	if manager.GetTargetURL() != newTarget {
		t.Errorf("Expected target URL to be %q, got %q", newTarget, manager.GetTargetURL())
	}

	// Test with invalid target URL
	if err := manager.UpdateTarget("://invalid-url"); err == nil {
		t.Fatal("Expected error for invalid target URL, got nil")
	}
}

// TestUpdatePathRewrite tests the UpdatePathRewrite function
func TestUpdatePathRewrite(t *testing.T) {
	cfg := createTestConfig()
	manager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Update the path rewrite rules
	newPathRewrite := map[string]string{
		"^/api/v1": "/v1",
		"^/api/v2": "/v2",
	}
	if err := manager.UpdatePathRewrite(newPathRewrite); err != nil {
		t.Fatalf("Failed to update path rewrite: %v", err)
	}

	// Check that the path rewrite rules were updated
	pathRewrite := manager.GetPathRewrite()
	if len(pathRewrite) != len(newPathRewrite) {
		t.Errorf("Expected %d path rewrite rules, got %d", len(newPathRewrite), len(pathRewrite))
	}
	for pattern, replacement := range newPathRewrite {
		if pathRewrite[pattern] != replacement {
			t.Errorf("Expected path rewrite rule %q -> %q, got %q", pattern, replacement, pathRewrite[pattern])
		}
	}
}

// TestSetChangeOrigin tests the SetChangeOrigin function
func TestSetChangeOrigin(t *testing.T) {
	cfg := createTestConfig()
	manager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Initial state should be true (from config)
	if !manager.IsChangeOrigin() {
		t.Error("Expected initial changeOrigin to be true")
	}

	// Set to false
	if err := manager.SetChangeOrigin(false); err != nil {
		t.Fatalf("Failed to set changeOrigin: %v", err)
	}

	// Check that it was updated
	if manager.IsChangeOrigin() {
		t.Error("Expected changeOrigin to be false after update")
	}

	// Set back to true
	if err := manager.SetChangeOrigin(true); err != nil {
		t.Fatalf("Failed to set changeOrigin: %v", err)
	}

	// Check that it was updated
	if !manager.IsChangeOrigin() {
		t.Error("Expected changeOrigin to be true after second update")
	}
}