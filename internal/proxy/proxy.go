package proxy

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/logger"
	"github.com/mockoho/mockoho/internal/middleware"
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

	// Add response modifier to remove CORS headers from the proxied response
	// to prevent duplicate headers when our middleware adds them
	originalModifyResponse := proxy.ModifyResponse
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove any existing CORS headers to prevent duplicates
		for header := range middleware.CORSHeaders {
			resp.Header.Del(header)
		}
		
		// Call the original modifier if it exists
		if originalModifyResponse != nil {
			return originalModifyResponse(resp)
		}
		return nil
	}

	// Add custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Special handling for aborted requests (like WebSockets)
		if err == http.ErrAbortHandler {
			return
		}
		
		logger.ProxyError(targetURL.String(), err)
		w.WriteHeader(http.StatusBadGateway)
		_, writeErr := w.Write([]byte("Proxy Error"))
		if writeErr != nil {
			logger.Error("Failed to write proxy error response: %v", writeErr)
		}
	}

	return proxy
}

// Handle handles a request by proxying it to the real server
func (m *Manager) Handle(c *gin.Context) {
	// Create a response recorder to capture the status code and response body
	responseRecorder := &responseRecorder{
		ResponseWriter: c.Writer,
		statusCode:     http.StatusOK, // Default status code
		written:        false,
		body:           make([]byte, 0, 1024), // Pre-allocate buffer with reasonable capacity
		headers:        make(http.Header),     // Initialize headers map
	}
	
	// Use the response recorder instead of the original writer
	start := time.Now()
	
	// Handle potential panics from the proxy
	defer func() {
		if err := recover(); err != nil {
			// Check if it's the special ErrAbortHandler which is expected in some cases
			if err == http.ErrAbortHandler {
				// Normal for WebSockets, no need to log
			} else {
				// Re-panic for other errors
				panic(err)
			}
		}
		
		// Only log if a response was actually written
		if responseRecorder.written {
			// Log the proxied request with method, path and status
			logger.Info("%s %s - proxied - %d (%s)",
				c.Request.Method,
				c.Request.URL.Path,
				responseRecorder.statusCode,
				time.Since(start))
		}
	}()
	
	// Create a custom transport that copies all headers
	originalTransport := m.proxy.Transport
	m.proxy.Transport = &headerCopyingTransport{
		originalTransport: originalTransport,
		responseRecorder: responseRecorder,
	}
	
	m.proxy.ServeHTTP(responseRecorder, c.Request)
	
	// Restore original transport
	m.proxy.Transport = originalTransport
}

// responseRecorder is a wrapper for http.ResponseWriter that captures the status code and response body
type responseRecorder struct {
	gin.ResponseWriter
	statusCode int
	written    bool
	body       []byte // Buffer to store the response body
	headers    http.Header // Store headers separately
}

// Header returns the header map that will be sent by WriteHeader
func (r *responseRecorder) Header() http.Header {
	return r.headers
}

// WriteHeader captures the status code before writing it
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.written = true
	
	// Copy all headers from our custom headers to the underlying ResponseWriter
	for key, values := range r.headers {
		for _, value := range values {
			r.ResponseWriter.Header().Set(key, value)
		}
	}
	
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write captures that the response has been written and stores the response body
func (r *responseRecorder) Write(b []byte) (int, error) {
	// Ensure headers are copied before writing the body if WriteHeader wasn't called
	if !r.written {
		// Copy all headers from our custom headers to the underlying ResponseWriter
		for key, values := range r.headers {
			for _, value := range values {
				r.ResponseWriter.Header().Set(key, value)
			}
		}
		r.written = true
	}
	
	// Store a copy of the response body (up to a reasonable size limit)
	if len(r.body) < 1024*1024 { // Limit to 1MB to prevent memory issues
		r.body = append(r.body, b...)
	}
	
	return r.ResponseWriter.Write(b)
}

// Hijack implements the http.Hijacker interface to support WebSocket
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("the ResponseWriter doesn't support hijacking")
}

// Flush implements the http.Flusher interface
func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// CloseNotify implements the http.CloseNotifier interface
// NOTE: This is deprecated in newer Go versions and should be replaced with context cancellation
// Kept for backward compatibility with older Go versions
func (r *responseRecorder) CloseNotify() <-chan bool {
	if closeNotifier, ok := r.ResponseWriter.(http.CloseNotifier); ok {
		return closeNotifier.CloseNotify()
	}
	return nil
}

// Push implements the http.Pusher interface
func (r *responseRecorder) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// UpdateTarget updates the proxy target
func (m *Manager) UpdateTarget(target string) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		logger.Error("Failed to parse target URL: %v", err)
		return err
	}

	m.Config.Global.ProxyConfig.Target = target
	
	// Create a new proxy with the updated target
	m.proxy = createReverseProxy(targetURL, m.Config)
	
	// Save the global config
	err = m.Config.SaveGlobalConfig()
	if err != nil {
		logger.Error("Failed to save global config: %v", err)
		return err
	}
	
	// Log success at info level
	logger.Info("Proxy target updated to: %s", target)
	
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


// headerCopyingTransport is a custom http.RoundTripper that ensures all headers
// from the server response are properly copied to our response
type headerCopyingTransport struct {
	originalTransport http.RoundTripper
	responseRecorder  *responseRecorder
}

// RoundTrip implements the http.RoundTripper interface
func (t *headerCopyingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Use the original transport if available, otherwise use the default transport
	transport := t.originalTransport
	if transport == nil {
		transport = http.DefaultTransport
	}
	
	// Perform the actual request
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	
	// Copy all headers from the server response to our responseRecorder
	// Skip CORS headers as they will be set by the middleware.CORSMiddleware
	for key, values := range resp.Header {
		// Skip CORS headers
		if middleware.CORSHeaders[key] {
			continue
		}
		
		for _, value := range values {
			// Use Set instead of Add to ensure we don't get duplicate headers
			t.responseRecorder.Header().Set(key, value)
		}
	}
	
	return resp, nil
}
