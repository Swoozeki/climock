package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kohofinancial/mockoho/internal/config"
	"kohofinancial/mockoho/internal/logger"
	"kohofinancial/mockoho/internal/middleware"
	"kohofinancial/mockoho/internal/mock"
	"kohofinancial/mockoho/internal/proxy"

	"github.com/gin-gonic/gin"
)

func init() {
	// Set Gin to release mode to disable debug logging
	gin.SetMode(gin.ReleaseMode)
}

// Server represents the mock server
type Server struct {
	Config      *config.Config
	MockManager *mock.Manager
	ProxyManager *proxy.Manager
	router      *gin.Engine
	httpServer  *http.Server
	isRunning   bool
}

// New creates a new server
func New(cfg *config.Config, mockManager *mock.Manager, proxyManager *proxy.Manager) *Server {
	server := &Server{
		Config:      cfg,
		MockManager: mockManager,
		ProxyManager: proxyManager,
		isRunning:   false,
	}
	
	// Initialize router
	server.setupRoutes()
	
	return server
}


// Start starts the server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.Config.Global.ServerConfig.Host, s.Config.Global.ServerConfig.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server started at %s", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Error starting server: %v", err)
		}
	}()

	s.isRunning = true
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	if !s.isRunning {
		return fmt.Errorf("server is not running")
	}

	logger.Info("Stopping server at %s:%d", s.Config.Global.ServerConfig.Host, s.Config.Global.ServerConfig.Port)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Error shutting down server: %v", err)
		return err
	}

	s.isRunning = false
	logger.Info("Server stopped")
	return nil
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	return s.isRunning
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.Config.Global.ServerConfig.Host, s.Config.Global.ServerConfig.Port)
}

// setupRoutes sets up the server routes
func (s *Server) setupRoutes() {
	// Create a new router
	s.router = gin.New()
	// Add recovery middleware
	s.router.Use(gin.Recovery())
	// Add CORS middleware
	s.router.Use(middleware.CORSMiddleware())

	// Add a catch-all route to handle all requests
	s.router.Any("/*path", s.handleRequest)
}

// handleRequest handles an incoming request
func (s *Server) handleRequest(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	// Try to find a matching endpoint
	endpoint, _, err := s.MockManager.FindEndpoint(method, path)
	if err != nil || !endpoint.Active {
		// No matching endpoint or endpoint is inactive, proxy the request
		s.ProxyManager.Handle(c)
		return
	}

	// Handle the mock response
	s.handleMockResponse(c, endpoint, path)
}

// handleMockResponse generates and sends a mock response
func (s *Server) handleMockResponse(c *gin.Context, endpoint *config.Endpoint, path string) {
	// Extract path parameters
	params := s.MockManager.ExtractParams(endpoint.Path, path)

	// Generate response
	response, err := s.MockManager.GenerateResponse(endpoint, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to generate response: %v", err),
		})
		return
	}

	// Apply delay if specified
	if response.Delay > 0 {
		time.Sleep(time.Duration(response.Delay) * time.Millisecond)
	}

	// Send the response
	s.sendResponse(c, response)
}

// setResponseHeaders sets the response headers
func (s *Server) setResponseHeaders(c *gin.Context, headers map[string]string) {
	// List of CORS headers that should not be overridden
	corsHeaders := middleware.CORSHeaders

	for key, value := range headers {
		// Skip CORS headers that are already set by the middleware
		if corsHeaders[key] {
			continue
		}
		c.Header(key, value)
	}
}

// writeStringJSONBody attempts to write a string as JSON response
// Returns true if the string was valid JSON and was written successfully
func (s *Server) writeStringJSONBody(c *gin.Context, bodyStr string) bool {
	var jsonBody interface{}
	if err := json.Unmarshal([]byte(bodyStr), &jsonBody); err == nil {
		c.Writer.Header().Set("Content-Type", "application/json")
		if _, err := c.Writer.WriteString(bodyStr); err != nil {
			logger.Error("Failed to write JSON response: %v", err)
		}
		return true
	}
	return false
}


// sendResponse sends the response to the client
func (s *Server) sendResponse(c *gin.Context, response *config.Response) {
	// Set response headers
	s.setResponseHeaders(c, response.Headers)

	// Set response status
	c.Status(response.Status)

	// Handle string JSON bodies
	if bodyStr, ok := response.Body.(string); ok {
		if s.writeStringJSONBody(c, bodyStr) {
			// Log the request
			start := time.Now()
			logger.Info("%s %s - mocked - %d (%s)",
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				time.Since(start))
			return
		}
	}

	// Otherwise, render as JSON
	c.JSON(response.Status, response.Body)

	// Log the request
	start := time.Now()
	logger.Info("%s %s - mocked - %d (%s)",
		c.Request.Method,
		c.Request.URL.Path,
		c.Writer.Status(),
		time.Since(start))
}

// Reload reloads the server configuration
func (s *Server) Reload() error {
	// Reload configuration
	if err := s.Config.Load(); err != nil {
		return err
	}

	// Update routes if the server is running
	if s.isRunning {
		s.setupRoutes()
	}

	return nil
}

// UpdatePort updates the server port
func (s *Server) UpdatePort(port int) error {
	if s.isRunning {
		return fmt.Errorf("cannot change port while server is running")
	}

	s.Config.Global.ServerConfig.Port = port
	return s.Config.SaveGlobalConfig()
}

// UpdateHost updates the server host
func (s *Server) UpdateHost(host string) error {
	if s.isRunning {
		return fmt.Errorf("cannot change host while server is running")
	}

	s.Config.Global.ServerConfig.Host = host
	return s.Config.SaveGlobalConfig()
}