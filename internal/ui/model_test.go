package ui_test

import (
	"testing"

	"kohofinancial/mockoho/internal/config"
	"kohofinancial/mockoho/internal/logger"
	"kohofinancial/mockoho/internal/mock"
	"kohofinancial/mockoho/internal/proxy"
	"kohofinancial/mockoho/internal/server"
	"kohofinancial/mockoho/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	// Initialize test logger to prevent nil pointer dereferences
	logger.InitTestLogger()
}

// createTestConfig creates a test configuration for UI tests
func createTestConfig() *config.Config {
	cfg := config.New("")

	// Set up global config
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
		Editor: config.EditorConfig{
			Command: "code",
			Args:    []string{"-g", "{file}:{line}"},
		},
	}

	// Set up a test feature with endpoints
	feature := config.FeatureConfig{
		Feature: "test",
		Endpoints: []config.Endpoint{
			{
				ID:              "endpoint1",
				Method:          "GET",
				Path:            "/api/test1",
				Active:          true,
				DefaultResponse: "standard",
				Responses: map[string]config.Response{
					"standard": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"message": "Test 1",
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
				ID:              "endpoint2",
				Method:          "GET",
				Path:            "/api/test2",
				Active:          false,
				DefaultResponse: "standard",
				Responses: map[string]config.Response{
					"standard": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]string{
							"message": "Test 2",
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

// TestNewModel tests the New function for creating a UI model
func TestNewModel(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)
	if model == nil {
		t.Fatal("Expected non-nil UI model")
	}

	// Check that the model was initialized with the correct components
	if model.Config != cfg {
		t.Error("Expected model.Config to be the provided config")
	}
	if model.MockManager != mockManager {
		t.Error("Expected model.MockManager to be the provided mock manager")
	}
	if model.ProxyManager != proxyManager {
		t.Error("Expected model.ProxyManager to be the provided proxy manager")
	}
	if model.Server != srv {
		t.Error("Expected model.Server to be the provided server")
	}
}

// TestModelInit tests the Init function of the UI model
func TestModelInit(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)

	// Call Init
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Expected non-nil command from Init")
	}
}

// TestModelUpdate tests the Update function of the UI model
func TestModelUpdate(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)

	// Test window size message
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if updatedModel == nil {
		t.Fatal("Expected non-nil model from Update")
	}

	// Test key message for tab (switch panel)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if updatedModel == nil {
		t.Fatal("Expected non-nil model from Update")
	}

	// Test key message for quit
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if updatedModel == nil {
		t.Fatal("Expected non-nil model from Update")
	}
}

// TestModelView tests the View function of the UI model
func TestModelView(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)

	// Call View
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

// TestDialogHandling tests the dialog handling functionality
func TestDialogHandling(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)

	// Test dialog handling through key messages
	// We can't directly test the dialog state as it's private,
	// but we can test that the model handles dialog-related key messages
	
	// First, simulate opening a dialog with a key message
	// (this is just testing the Update method handles the message without crashing)
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}) // 'n' for new
	
	// Then simulate ESC to cancel the dialog
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	
	// The model should still be valid and render a view
	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view after dialog interaction")
	}
}

// TestKeyHandling tests the key handling functionality
func TestKeyHandling(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)

	// Test various key messages
	keyTests := []struct {
		name    string
		key     tea.KeyType
		runes   []rune
		alt     bool
		wantCmd bool
	}{
		{"Tab", tea.KeyTab, nil, false, true},
		{"Up", tea.KeyUp, nil, false, true},
		{"Down", tea.KeyDown, nil, false, true},
		{"Enter", tea.KeyEnter, nil, false, true},
		{"t (toggle)", tea.KeyRunes, []rune{'t'}, false, true},
		{"r (response)", tea.KeyRunes, []rune{'r'}, false, true},
		{"s (server)", tea.KeyRunes, []rune{'s'}, false, true},
		{"q (quit)", tea.KeyRunes, []rune{'q'}, false, true},
		{"h (help)", tea.KeyRunes, []rune{'h'}, false, true},
	}

	for _, tt := range keyTests {
		t.Run(tt.name, func(t *testing.T) {
			var keyMsg tea.KeyMsg
			if tt.runes != nil {
				keyMsg = tea.KeyMsg{Type: tt.key, Runes: tt.runes, Alt: tt.alt}
			} else {
				keyMsg = tea.KeyMsg{Type: tt.key, Alt: tt.alt}
			}
			
			_, _ = model.Update(keyMsg)
			// We don't check the command as it's implementation-dependent
		})
	}
}

// TestServerInteraction tests the server interaction functionality
func TestServerInteraction(t *testing.T) {
	cfg := createTestConfig()
	mockManager := mock.New(cfg)
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}
	srv := server.New(cfg, mockManager, proxyManager)

	// Create a new UI model
	model := ui.New(cfg, mockManager, proxyManager, srv)

	// In a test environment, we can't actually start the server
	// So we just verify that the model handles the key press without crashing
	_, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	
	// We can't reliably test the server state in a unit test
	// as it depends on network resources
}