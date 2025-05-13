package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/mockoho/mockoho/internal/config"
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

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Customize the director function to modify the request
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

	return &Manager{
		Config: cfg,
		proxy:  proxy,
	}, nil
}

// Handle handles a request by proxying it to the real server
func (m *Manager) Handle(c *gin.Context) {
	m.proxy.ServeHTTP(c.Writer, c.Request)
}

// UpdateTarget updates the proxy target
func (m *Manager) UpdateTarget(target string) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		return err
	}

	m.Config.Global.ProxyConfig.Target = target
	
	// Create a new proxy with the updated target
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Customize the director function to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Apply path rewriting
		for pattern, replacement := range m.Config.Global.ProxyConfig.PathRewrite {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			req.URL.Path = re.ReplaceAllString(req.URL.Path, replacement)
		}

		// Set the Host header to the target host if changeOrigin is true
		if m.Config.Global.ProxyConfig.ChangeOrigin {
			req.Host = targetURL.Host
		}
	}

	m.proxy = proxy
	
	return m.Config.SaveGlobalConfig()
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