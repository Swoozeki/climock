package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/mock"
	"github.com/mockoho/mockoho/internal/proxy"
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
	// Use gin.New() instead of gin.Default() to avoid debug logging
	router := gin.New()
	// Add only the recovery middleware
	router.Use(gin.Recovery())
	
	return &Server{
		Config:      cfg,
		MockManager: mockManager,
		ProxyManager: proxyManager,
		router:      router,
		isRunning:   false,
	}
}

// Start starts the server
func (s *Server) Start() error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Set up routes
	s.setupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.Config.Global.ServerConfig.Host, s.Config.Global.ServerConfig.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server: %v\n", err)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	s.isRunning = false
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
	// Clear existing routes
	s.router = gin.New()
	// Add only the recovery middleware
	s.router.Use(gin.Recovery())

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

	// Set response headers
	for key, value := range response.Headers {
		c.Header(key, value)
	}

	// Set response status and body
	c.Status(response.Status)
	
	// Check if the body is a string that needs to be parsed as JSON
	if bodyStr, ok := response.Body.(string); ok {
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(bodyStr), &jsonBody); err == nil {
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteString(bodyStr)
			return
		}
	}
	
	// Otherwise, render as JSON
	c.JSON(response.Status, response.Body)
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