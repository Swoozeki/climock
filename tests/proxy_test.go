package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/proxy"
)

func TestProxyPathRewrite(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Global: config.GlobalConfig{
			ProxyConfig: config.ProxyConfig{
				Target:       "http://example.com",
				ChangeOrigin: true,
				PathRewrite: map[string]string{
					"^/api": "",
				},
			},
		},
	}

	// Create a test server that will verify the path rewrite
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The path should be rewritten from /api/users to /users
		if r.URL.Path != "/users" {
			t.Errorf("Expected path '/users', got '%s'", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Update the target to the test server
	cfg.Global.ProxyConfig.Target = testServer.URL

	// Create proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Create a test request to /api/users
	req, _ := http.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Handle the request
	proxyManager.Handle(c)

	// The test server will verify the path rewrite
}

func TestProxyChangeOrigin(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Global: config.GlobalConfig{
			ProxyConfig: config.ProxyConfig{
				Target:       "http://example.com",
				ChangeOrigin: true,
				PathRewrite:  map[string]string{},
			},
		},
	}

	// Create a test server that will verify the Host header
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The Host header should be set to the target host
		expectedHost := r.URL.Host
		if r.Host != expectedHost {
			t.Errorf("Expected Host header '%s', got '%s'", expectedHost, r.Host)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Update the target to the test server
	cfg.Global.ProxyConfig.Target = testServer.URL

	// Create proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Host = "original-host.com"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Handle the request
	proxyManager.Handle(c)

	// The test server will verify the Host header
}

func TestProxyUpdateTarget(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Global: config.GlobalConfig{
			ProxyConfig: config.ProxyConfig{
				Target:       "http://example.com",
				ChangeOrigin: true,
				PathRewrite:  map[string]string{},
			},
		},
	}

	// Create proxy manager
	proxyManager, err := proxy.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create proxy manager: %v", err)
	}

	// Update the target
	newTarget := "http://new-target.com"
	err = proxyManager.UpdateTarget(newTarget)
	if err != nil {
		t.Errorf("Failed to update target: %v", err)
	}

	// Verify the target was updated
	if proxyManager.GetTargetURL() != newTarget {
		t.Errorf("Expected target '%s', got '%s'", newTarget, proxyManager.GetTargetURL())
	}
}