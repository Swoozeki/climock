package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/mockoho/mockoho/internal/config"
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
		return nil, fmt.Errorf("response %s not found for endpoint %s", responseName, endpoint.ID)
	}

	// Process template variables in the response body
	processedResponse := response
	if err := m.processResponseBody(&processedResponse, params); err != nil {
		return nil, err
	}

	return &processedResponse, nil
}

// processResponseBody processes template variables in the response body
func (m *Manager) processResponseBody(response *config.Response, params map[string]string) error {
	// Convert body to JSON string
	bodyJSON, err := json.Marshal(response.Body)
	if err != nil {
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
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	// Parse the processed JSON back into the response body
	var processedBody interface{}
	if err := json.Unmarshal(buf.Bytes(), &processedBody); err != nil {
		return err
	}

	response.Body = processedBody
	return nil
}

// ToggleEndpoint toggles an endpoint's active state
func (m *Manager) ToggleEndpoint(feature, id string) error {
	endpoint, err := m.Config.GetEndpoint(feature, id)
	if err != nil {
		return err
	}

	endpoint.Active = !endpoint.Active
	if err := m.Config.UpdateEndpoint(feature, *endpoint); err != nil {
		return err
	}

	return m.Config.SaveFeatureConfig(feature)
}

// SetDefaultResponse sets the default response for an endpoint
func (m *Manager) SetDefaultResponse(feature, id, response string) error {
	endpoint, err := m.Config.GetEndpoint(feature, id)
	if err != nil {
		return err
	}

	if _, ok := endpoint.Responses[response]; !ok {
		return fmt.Errorf("response %s not found for endpoint %s", response, id)
	}

	endpoint.DefaultResponse = response
	if err := m.Config.UpdateEndpoint(feature, *endpoint); err != nil {
		return err
	}

	return m.Config.SaveFeatureConfig(feature)
}

// CreateEndpoint creates a new endpoint
func (m *Manager) CreateEndpoint(feature string, endpoint config.Endpoint) error {
	fmt.Printf("Mock Manager: Creating endpoint %s in feature %s\n", endpoint.ID, feature)
	
	if err := m.Config.AddEndpoint(feature, endpoint); err != nil {
		fmt.Printf("Error adding endpoint to config: %v\n", err)
		return fmt.Errorf("failed to add endpoint to config: %w", err)
	}

	fmt.Printf("Endpoint added to in-memory config, saving to file...\n")
	if err := m.Config.SaveFeatureConfig(feature); err != nil {
		fmt.Printf("Error saving feature config: %v\n", err)
		return fmt.Errorf("failed to save feature config: %w", err)
	}
	
	fmt.Printf("Feature config saved successfully\n")
	return nil
}

// CreateFeature creates a new feature
func (m *Manager) CreateFeature(feature config.FeatureConfig) error {
	if err := m.Config.AddFeature(feature); err != nil {
		return fmt.Errorf("failed to add feature to config: %w", err)
	}

	if err := m.Config.SaveFeatureConfig(feature.Feature); err != nil {
		return fmt.Errorf("failed to save feature config: %w", err)
	}
	
	return nil
}

// DeleteEndpoint deletes an endpoint
func (m *Manager) DeleteEndpoint(feature, id string) error {
	if err := m.Config.DeleteEndpoint(feature, id); err != nil {
		return err
	}

	return m.Config.SaveFeatureConfig(feature)
}

// DeleteFeature deletes a feature
func (m *Manager) DeleteFeature(feature string) error {
	return m.Config.DeleteFeature(feature)
}