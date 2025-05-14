package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/logger"
)

// Manager handles mock endpoints and response generation
type Manager struct {
	Config *config.Config
}

// New creates a new mock manager
func New(cfg *config.Config) *Manager {
	return &Manager{
		Config: cfg,
	}
}

// FindEndpoint finds an endpoint matching the given method and path
func (m *Manager) FindEndpoint(method, path string) (*config.Endpoint, string, error) {
	for feature, featureConfig := range m.Config.Mocks {
		for i := range featureConfig.Endpoints {
			endpoint := &featureConfig.Endpoints[i]
			if endpoint.Method == method && m.pathMatches(endpoint.Path, path) {
				return endpoint, feature, nil
			}
		}
	}
	logger.LogDebug("No matching endpoint found for %s %s", method, path)
	return nil, "", fmt.Errorf("no matching endpoint found for %s %s", method, path)
}

// pathMatches checks if a request path matches an endpoint path pattern
func (m *Manager) pathMatches(pattern, path string) bool {
	// Convert pattern to regex
	parts := strings.Split(pattern, "/")
	regexParts := make([]string, len(parts))

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			// This is a parameter
			regexParts[i] = "[^/]+"
		} else {
			regexParts[i] = part
		}
	}

	// Simple path matching for now
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			// This is a parameter, so it matches anything
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}

	return true
}

// ExtractParams extracts path parameters from a request path
func (m *Manager) ExtractParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			paramName := patternParts[i][1:] // Remove the : prefix
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// GenerateResponse generates a response for the given endpoint and parameters
func (m *Manager) GenerateResponse(endpoint *config.Endpoint, params map[string]string) (*config.Response, error) {
	responseName := endpoint.DefaultResponse
	response, ok := endpoint.Responses[responseName]
	if !ok {
		logger.Error("Response %s not found for endpoint %s", responseName, endpoint.ID)
		return nil, fmt.Errorf("response %s not found for endpoint %s", responseName, endpoint.ID)
	}

	// Process template variables in the response body
	processedResponse := response
	if err := m.processResponseBody(&processedResponse, params); err != nil {
		logger.Error("Failed to process response body: %v", err)
		return nil, err
	}

	return &processedResponse, nil
}

// processResponseBody processes template variables in the response body
func (m *Manager) processResponseBody(response *config.Response, params map[string]string) error {
	// Convert body to JSON string
	bodyJSON, err := json.Marshal(response.Body)
	if err != nil {
		logger.Error("Failed to marshal response body: %v", err)
		return err
	}

	// Create template data
	data := map[string]interface{}{
		"params": params,
		"now":    time.Now().Format(time.RFC3339),
	}

	// Process template
	tmpl, err := template.New("body").Parse(string(bodyJSON))
	if err != nil {
		logger.Error("Failed to parse response template: %v", err)
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.Error("Failed to execute response template: %v", err)
		return err
	}

	// Parse the processed JSON back into the response body
	var processedBody interface{}
	if err := json.Unmarshal(buf.Bytes(), &processedBody); err != nil {
		logger.Error("Failed to unmarshal processed response: %v", err)
		return err
	}

	response.Body = processedBody
	return nil
}

// ToggleEndpoint toggles an endpoint's active state
func (m *Manager) ToggleEndpoint(feature, id string) error {
	endpoint, err := m.Config.GetEndpoint(feature, id)
	if err != nil {
		logger.Error("Failed to get endpoint %s in feature %s: %v", id, feature, err)
		return err
	}

	endpoint.Active = !endpoint.Active
	if err := m.Config.UpdateEndpoint(feature, *endpoint); err != nil {
		logger.Error("Failed to update endpoint %s in feature %s: %v", id, feature, err)
		return err
	}
	
	logger.Info("Toggled endpoint %s in feature %s to %v", id, feature, endpoint.Active)

	return m.Config.SaveFeatureConfig(feature)
}

// SetDefaultResponse sets the default response for an endpoint
func (m *Manager) SetDefaultResponse(feature, id, response string) error {
	endpoint, err := m.Config.GetEndpoint(feature, id)
	if err != nil {
		logger.Error("Failed to get endpoint %s in feature %s: %v", id, feature, err)
		return err
	}

	if _, ok := endpoint.Responses[response]; !ok {
		logger.Error("Response %s not found for endpoint %s", response, id)
		return fmt.Errorf("response %s not found for endpoint %s", response, id)
	}

	endpoint.DefaultResponse = response
	if err := m.Config.UpdateEndpoint(feature, *endpoint); err != nil {
		logger.Error("Failed to update endpoint %s in feature %s: %v", id, feature, err)
		return err
	}
	
	logger.Info("Set default response for endpoint %s in feature %s to %s", id, feature, response)

	return m.Config.SaveFeatureConfig(feature)
}

// CreateEndpoint creates a new endpoint
func (m *Manager) CreateEndpoint(feature string, endpoint config.Endpoint) error {
	logger.Info("Creating endpoint %s in feature %s", endpoint.ID, feature)
	
	if err := m.Config.AddEndpoint(feature, endpoint); err != nil {
		logger.Error("Failed to add endpoint to config: %v", err)
		return fmt.Errorf("failed to add endpoint to config: %w", err)
	}

	logger.LogDebug("Endpoint added to in-memory config, saving to file...")
	if err := m.Config.SaveFeatureConfig(feature); err != nil {
		logger.Error("Failed to save feature config: %v", err)
		return fmt.Errorf("failed to save feature config: %w", err)
	}
	
	logger.Info("Endpoint %s created successfully in feature %s", endpoint.ID, feature)
	return nil
}

// CreateFeature creates a new feature
func (m *Manager) CreateFeature(feature config.FeatureConfig) error {
	logger.Info("Creating feature %s", feature.Feature)
	
	if err := m.Config.AddFeature(feature); err != nil {
		logger.Error("Failed to add feature to config: %v", err)
		return fmt.Errorf("failed to add feature to config: %w", err)
	}

	if err := m.Config.SaveFeatureConfig(feature.Feature); err != nil {
		logger.Error("Failed to save feature config: %v", err)
		return fmt.Errorf("failed to save feature config: %w", err)
	}
	
	logger.Info("Feature %s created successfully", feature.Feature)
	return nil
}

// DeleteEndpoint deletes an endpoint
func (m *Manager) DeleteEndpoint(feature, id string) error {
	logger.Info("Deleting endpoint %s from feature %s", id, feature)
	
	if err := m.Config.DeleteEndpoint(feature, id); err != nil {
		logger.Error("Failed to delete endpoint %s from feature %s: %v", id, feature, err)
		return err
	}

	if err := m.Config.SaveFeatureConfig(feature); err != nil {
		logger.Error("Failed to save feature config after deleting endpoint: %v", err)
		return err
	}
	
	logger.Info("Endpoint %s deleted successfully from feature %s", id, feature)
	return nil
}

// DeleteFeature deletes a feature
func (m *Manager) DeleteFeature(feature string) error {
	logger.Info("Deleting feature %s", feature)
	
	if err := m.Config.DeleteFeature(feature); err != nil {
		logger.Error("Failed to delete feature %s: %v", feature, err)
		return err
	}
	
	logger.Info("Feature %s deleted successfully", feature)
	return nil
}