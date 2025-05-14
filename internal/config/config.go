package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mockoho/mockoho/internal/logger"
)

// ProxyConfig holds the proxy server configuration
type ProxyConfig struct {
	Target       string            `json:"target"`
	ChangeOrigin bool              `json:"changeOrigin"`
	PathRewrite  map[string]string `json:"pathRewrite"`
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// EditorConfig holds the external editor configuration
type EditorConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// GlobalConfig holds the global application configuration
type GlobalConfig struct {
	ProxyConfig  ProxyConfig  `json:"proxyConfig"`
	ServerConfig ServerConfig `json:"serverConfig"`
	Editor       EditorConfig `json:"editor"`
}

// Config holds the entire application configuration
type Config struct {
	Global  GlobalConfig
	Mocks   map[string]FeatureConfig
	BaseDir string
	mu      sync.RWMutex
}

// FeatureConfig holds the configuration for a specific feature
type FeatureConfig struct {
	Feature   string     `json:"feature"`
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint represents a mock API endpoint
type Endpoint struct {
	ID              string              `json:"id"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	Active          bool                `json:"active"`
	DefaultResponse string              `json:"defaultResponse"`
	Responses       map[string]Response `json:"responses"`
}

// Response represents a mock API response
type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Delay   int               `json:"delay"`
}

// New creates a new Config instance
func New(baseDir string) *Config {
	return &Config{
		Mocks:   make(map[string]FeatureConfig),
		BaseDir: baseDir,
	}
}

// Load loads the configuration from the specified directory
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Load global config
	globalConfigPath := filepath.Join(c.BaseDir, "config.json")
	if err := c.loadGlobalConfig(globalConfigPath); err != nil {
		logger.Error("Failed to load global config: %v", err)
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Load feature configs
	files, err := os.ReadDir(c.BaseDir)
	if err != nil {
		logger.Error("Failed to read mocks directory: %v", err)
		return fmt.Errorf("failed to read mocks directory: %w", err)
	}

	c.Mocks = make(map[string]FeatureConfig)
	for _, file := range files {
		if file.IsDir() || file.Name() == "config.json" {
			continue
		}

		featurePath := filepath.Join(c.BaseDir, file.Name())
		featureConfig, err := c.loadFeatureConfig(featurePath)
		if err != nil {
			logger.Error("Failed to load feature config %s: %v", file.Name(), err)
			return fmt.Errorf("failed to load feature config %s: %w", file.Name(), err)
		}

		c.Mocks[featureConfig.Feature] = featureConfig
	}

	return nil
}

// loadGlobalConfig loads the global configuration from the specified file
func (c *Config) loadGlobalConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &c.Global); err != nil {
		return err
	}

	return nil
}

// loadFeatureConfig loads a feature configuration from the specified file
func (c *Config) loadFeatureConfig(path string) (FeatureConfig, error) {
	var config FeatureConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}

// SaveFeatureConfig saves a feature configuration to its file
func (c *Config) SaveFeatureConfig(feature string) error {
	c.mu.RLock()
	featureConfig, ok := c.Mocks[feature]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("feature %s not found", feature)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.BaseDir, feature+".json")
	
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create directory: %v", err)
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	data, err := json.MarshalIndent(featureConfig, "", "  ")
	if err != nil {
		return err
	}
	
	// Create a temporary file in the same directory
	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		logger.Error("Failed to write temporary file: %v", err)
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	
	// Rename the temporary file to the target file (atomic operation)
	if err := os.Rename(tempFile, path); err != nil {
		// Try to remove the temporary file
		os.Remove(tempFile)
		logger.Error("Failed to rename temporary file: %v", err)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}
	
	logger.Info("Saved feature config: %s", path)
	
	return nil
}

// SaveGlobalConfig saves the global configuration to its file
func (c *Config) SaveGlobalConfig() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	path := filepath.Join(c.BaseDir, "config.json")
	data, err := json.MarshalIndent(c.Global, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetEndpoint returns an endpoint by its ID
func (c *Config) GetEndpoint(feature, id string) (*Endpoint, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	featureConfig, ok := c.Mocks[feature]
	if !ok {
		return nil, fmt.Errorf("feature %s not found", feature)
	}

	for i := range featureConfig.Endpoints {
		if featureConfig.Endpoints[i].ID == id {
			return &featureConfig.Endpoints[i], nil
		}
	}

	return nil, fmt.Errorf("endpoint %s not found in feature %s", id, feature)
}

// UpdateEndpoint updates an endpoint
func (c *Config) UpdateEndpoint(feature string, endpoint Endpoint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	featureConfig, ok := c.Mocks[feature]
	if !ok {
		return fmt.Errorf("feature %s not found", feature)
	}

	for i := range featureConfig.Endpoints {
		if featureConfig.Endpoints[i].ID == endpoint.ID {
			featureConfig.Endpoints[i] = endpoint
			c.Mocks[feature] = featureConfig
			return nil
		}
	}

	return fmt.Errorf("endpoint %s not found in feature %s", endpoint.ID, feature)
}

// AddEndpoint adds a new endpoint to a feature
func (c *Config) AddEndpoint(feature string, endpoint Endpoint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	featureConfig, ok := c.Mocks[feature]
	if !ok {
		return fmt.Errorf("feature %s not found", feature)
	}

	// Check if endpoint with same ID already exists
	for _, e := range featureConfig.Endpoints {
		if e.ID == endpoint.ID {
			return fmt.Errorf("endpoint with ID %s already exists in feature %s", endpoint.ID, feature)
		}
	}

	featureConfig.Endpoints = append(featureConfig.Endpoints, endpoint)
	c.Mocks[feature] = featureConfig
	return nil
}

// AddFeature adds a new feature
func (c *Config) AddFeature(feature FeatureConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.Mocks[feature.Feature]; ok {
		return fmt.Errorf("feature %s already exists", feature.Feature)
	}

	c.Mocks[feature.Feature] = feature
	return nil
}

// DeleteEndpoint deletes an endpoint from a feature
func (c *Config) DeleteEndpoint(feature, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	featureConfig, ok := c.Mocks[feature]
	if !ok {
		return fmt.Errorf("feature %s not found", feature)
	}

	for i := range featureConfig.Endpoints {
		if featureConfig.Endpoints[i].ID == id {
			// Remove the endpoint
			featureConfig.Endpoints = append(
				featureConfig.Endpoints[:i],
				featureConfig.Endpoints[i+1:]...,
			)
			c.Mocks[feature] = featureConfig
			return nil
		}
	}

	return fmt.Errorf("endpoint %s not found in feature %s", id, feature)
}

// DeleteFeature deletes a feature
func (c *Config) DeleteFeature(feature string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.Mocks[feature]; !ok {
		return fmt.Errorf("feature %s not found", feature)
	}

	delete(c.Mocks, feature)
	
	// Delete the feature file
	path := filepath.Join(c.BaseDir, feature+".json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		logger.Error("Error removing feature file %s: %v", path, err)
		return fmt.Errorf("failed to remove feature file: %w", err)
	}
	
	logger.Info("Feature %s deleted successfully", feature)
	return nil
}