package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/logger"
)

// Manager handles proxying requests to the real server
type Manager struct {
	Config *config.Config
	proxy  *httputil.ReverseProxy
}

// New creates a new proxy manager
func New(cfg *config.Config) (*Manager, error) {
	targetURL, err := url.Parse(cfg.Global.ProxyConfig.Target)
	if err != nil {
		return nil, err
	}

	proxy := createReverseProxy(targetURL, cfg)

	return &Manager{
		Config: cfg,
		proxy:  proxy,
	}, nil
}

// createReverseProxy creates a configured reverse proxy for the given target URL
func createReverseProxy(targetURL *url.URL, cfg *config.Config) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Configure director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Apply path rewriting
		for pattern, replacement := range cfg.Global.ProxyConfig.PathRewrite {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			req.URL.Path = re.ReplaceAllString(req.URL.Path, replacement)
		}

		// Set the Host header to the target host if changeOrigin is true
		if cfg.Global.ProxyConfig.ChangeOrigin {
			req.Host = targetURL.Host
		}
	}

	// Add custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.ProxyError(targetURL.String(), err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Proxy Error"))
	}

	return proxy
}

// Handle handles a request by proxying it to the real server
func (m *Manager) Handle(c *gin.Context) {
	// Create a response recorder to capture the status code
	responseRecorder := &responseRecorder{
		ResponseWriter: c.Writer,
		statusCode:     http.StatusOK, // Default status code
	}
	
	// Use the response recorder instead of the original writer
	start := time.Now()
	m.proxy.ServeHTTP(responseRecorder, c.Request)
	
	// Log the proxied request
	logger.Info("Proxy response from %s to %s - %d (%s)",
		m.Config.Global.ProxyConfig.Target,
		c.Request.URL.Path,
		responseRecorder.statusCode,
		time.Since(start))
}

// responseRecorder is a wrapper for http.ResponseWriter that captures the status code
type responseRecorder struct {
	gin.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// UpdateTarget updates the proxy target
func (m *Manager) UpdateTarget(target string) error {
	logger.Info("Updating proxy target to: %s", target)
	
	targetURL, err := url.Parse(target)
	if err != nil {
		logger.Error("Failed to parse target URL: %v", err)
		return err
	}

	m.Config.Global.ProxyConfig.Target = target
	logger.Info("Set proxy target in config: %s", target)
	
	// Create a new proxy with the updated target
	m.proxy = createReverseProxy(targetURL, m.Config)
	logger.Info("Created new proxy with target: %s", target)
	
	logger.Info("Saving global config...")
	err = m.Config.SaveGlobalConfig()
	if err != nil {
		logger.Error("Failed to save global config: %v", err)
		return err
	}
	logger.Info("Global config saved successfully")
	
	return nil
}

// UpdatePathRewrite updates the path rewrite rules
func (m *Manager) UpdatePathRewrite(pathRewrite map[string]string) error {
	m.Config.Global.ProxyConfig.PathRewrite = pathRewrite
	return m.Config.SaveGlobalConfig()
}

// GetTargetURL returns the current proxy target URL
func (m *Manager) GetTargetURL() string {
	return m.Config.Global.ProxyConfig.Target
}

// GetPathRewrite returns the current path rewrite rules
func (m *Manager) GetPathRewrite() map[string]string {
	return m.Config.Global.ProxyConfig.PathRewrite
}

// IsChangeOrigin returns whether the proxy changes the origin
func (m *Manager) IsChangeOrigin() bool {
	return m.Config.Global.ProxyConfig.ChangeOrigin
}

// SetChangeOrigin sets whether the proxy changes the origin
func (m *Manager) SetChangeOrigin(changeOrigin bool) error {
	m.Config.Global.ProxyConfig.ChangeOrigin = changeOrigin
	return m.Config.SaveGlobalConfig()
}