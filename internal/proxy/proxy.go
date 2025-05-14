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
		// Log all headers from the server response for debugging
		if logger.IsDebugMode {
			logger.LogDebug("Headers in ModifyResponse (before modification):")
			for key, values := range resp.Header {
				for _, value := range values {
					logger.LogDebug("  %s: %s", key, value)
				}
			}
		}
		
		// Remove any existing CORS headers to prevent duplicates
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Expose-Headers")
		
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
			logger.LogDebug("HTTP connection closed by client or hijacked (normal for WebSockets)")
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
				logger.LogDebug("HTTP connection closed by client or hijacked (normal for WebSockets)")
			} else {
				// Re-panic for other errors
				panic(err)
			}
		}
		
		
		// Only log if a response was actually written
		if responseRecorder.written {
			// Log the proxied request
			logger.Info("Proxy response from %s to %s - %d (%s)",
				m.Config.Global.ProxyConfig.Target,
				c.Request.URL.Path,
				responseRecorder.statusCode,
				time.Since(start))
			
			// Log the response body in debug mode
			if logger.IsDebugMode {
				// Check content type to handle binary data appropriately
				contentType := responseRecorder.Header().Get("Content-Type")
				bodySize := len(responseRecorder.body)
				maxLogSize := 4096 // Limit log size to 4KB
				
				// Log the content type for debugging
				logger.LogDebug("Proxy response from %s has Content-Type: %s",
					m.Config.Global.ProxyConfig.Target,
					contentType)
				
				if bodySize > 0 {
					// Determine if this is likely binary data by checking both content type and content
					isBinary := isBinaryContent(contentType, responseRecorder.body)
					
					if isBinary {
						// For binary data, just log the content type and size
						logger.LogDebug("Proxy response body from %s: [Binary data of type %s, %d bytes]",
							m.Config.Global.ProxyConfig.Target,
							contentType,
							bodySize)
						
						// Log first few bytes as hex for debugging
						maxHexBytes := 32
						if bodySize < maxHexBytes {
							maxHexBytes = bodySize
						}
						hexStr := ""
						for i := 0; i < maxHexBytes; i++ {
							hexStr += fmt.Sprintf("%02x ", responseRecorder.body[i])
						}
						logger.LogDebug("First %d bytes (hex): %s", maxHexBytes, hexStr)
					} else {
						// For text data, log the actual content (with truncation if needed)
						if bodySize <= maxLogSize {
							logger.LogDebug("Proxy response body from %s (%s):\n%s",
								m.Config.Global.ProxyConfig.Target,
								contentType,
								string(responseRecorder.body))
						} else {
							// Truncate and indicate truncation
							logger.LogDebug("Proxy response body from %s (%s, truncated, %d bytes total):\n%s...",
								m.Config.Global.ProxyConfig.Target,
								contentType,
								bodySize,
								string(responseRecorder.body[:maxLogSize]))
						}
					}
				} else {
					logger.LogDebug("Proxy response from %s had empty body", m.Config.Global.ProxyConfig.Target)
				}
			}
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
	
	// Log headers for debugging
	if logger.IsDebugMode {
		logger.LogDebug("Sending headers to client:")
		for key, values := range r.ResponseWriter.Header() {
			for _, value := range values {
				logger.LogDebug("  %s: %s", key, value)
			}
		}
	}
	
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write captures that the response has been written and stores the response body
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.written = true
	// Store a copy of the response body (up to a reasonable size limit)
	if len(r.body) < 1024*1024 { // Limit to 1MB to prevent memory issues
		r.body = append(r.body, b...)
	}
	
	// Ensure headers are copied before writing the body if WriteHeader wasn't called
	if !r.written {
		// Copy all headers from our custom headers to the underlying ResponseWriter
		for key, values := range r.headers {
			for _, value := range values {
				r.ResponseWriter.Header().Set(key, value)
			}
		}
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
	// Only log at debug level for detailed operations
	logger.LogDebug("Updating proxy target to: %s", target)
	
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

// isBinaryContent determines if a content type or content represents binary data
func isBinaryContent(contentType string, content []byte) bool {
	// First check by content type
	if contentType != "" {
		// List of common binary content types
		binaryTypes := []string{
			"application/octet-stream",
			"application/pdf",
			"application/zip",
			"application/gzip",
			"application/x-gzip",
			"application/x-compressed",
			"application/x-zip-compressed",
			"image/",
			"audio/",
			"video/",
			"application/x-msdownload",
			"application/vnd.ms-",
			"application/vnd.openxmlformats-",
		}
		
		// Check if the content type matches any binary type
		for _, binaryType := range binaryTypes {
			if len(contentType) >= len(binaryType) && contentType[:len(binaryType)] == binaryType {
				return true
			}
		}
		
		// Check for compression encoding
		if contentType == "application/x-deflate" ||
		   contentType == "application/x-gzip" ||
		   contentType == "application/x-bzip2" {
			return true
		}
	}
	
	// If content type check didn't determine it's binary, check the content itself
	if len(content) > 0 {
		// Check for common binary signatures/magic numbers
		if len(content) >= 4 {
			// Check for gzip signature
			if content[0] == 0x1F && content[1] == 0x8B {
				return true
			}
			
			// Check for zip signature
			if content[0] == 0x50 && content[1] == 0x4B && content[2] == 0x03 && content[3] == 0x04 {
				return true
			}
			
			// Check for PDF signature
			if len(content) >= 5 && content[0] == 0x25 && content[1] == 0x50 && content[2] == 0x44 && content[3] == 0x46 {
				return true
			}
		}
		
		// Heuristic: Check if the content contains a high percentage of non-printable characters
		nonPrintable := 0
		sampleSize := 100
		if len(content) < sampleSize {
			sampleSize = len(content)
		}
		
		for i := 0; i < sampleSize; i++ {
			c := content[i]
			if (c < 32 || c > 126) && c != 9 && c != 10 && c != 13 { // Not printable ASCII and not tab, LF, CR
				nonPrintable++
			}
		}
		
		// If more than 20% of characters are non-printable, consider it binary
		if float64(nonPrintable)/float64(sampleSize) > 0.2 {
			return true
		}
	}
	
	return false
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
	
	// Log headers from server for debugging
	if logger.IsDebugMode {
		logger.LogDebug("Headers received from server:")
		for key, values := range resp.Header {
			for _, value := range values {
				logger.LogDebug("  %s: %s", key, value)
			}
		}
	}
	
	// Copy all headers from the server response to our responseRecorder
	// Skip CORS headers as they will be set by the corsMiddleware
	corsHeaders := map[string]bool{
		"Access-Control-Allow-Origin":      true,
		"Access-Control-Allow-Methods":     true,
		"Access-Control-Allow-Headers":     true,
		"Access-Control-Allow-Credentials": true,
		"Access-Control-Expose-Headers":    true,
	}
	
	for key, values := range resp.Header {
		// Skip CORS headers
		if corsHeaders[key] {
			if logger.IsDebugMode {
				logger.LogDebug("Skipping CORS header %s from proxied response", key)
			}
			continue
		}
		
		for _, value := range values {
			// Use Set instead of Add to ensure we don't get duplicate headers
			t.responseRecorder.Header().Set(key, value)
		}
	}
	
	return resp, nil
}
