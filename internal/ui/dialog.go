package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockoho/mockoho/internal/config"
)

// showNewFeatureDialog shows the new feature dialog
func (m *Model) showNewFeatureDialog() {
	// Clear any existing dialog state
	m.textInputs = nil
	m.dialogConfirmFn = nil
	m.dialogCancelFn = nil
	
	// Set dialog properties
	m.activeDialog = NewFeatureDialog
	m.dialogTitle = "Create New Feature"
	m.dialogContent = ""
	
	// Create text input for feature name with consistent styling
	ti := textinput.New()
	ti.Placeholder = "Feature name (letters, numbers, hyphens, underscores)"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 50
	
	// Store the text input in the model
	m.textInputs = []textinput.Model{ti}
	
	// No need to capture the value here, we'll get it directly from m.textInputs when needed
	
	// Set the confirm function - this will be called when Enter is pressed
	m.dialogConfirmFn = func() tea.Cmd {
		// Capture the feature name value now, before text inputs are cleared
		var featureName string
		if len(m.textInputs) > 0 {
			featureName = m.textInputs[0].Value()
		}
		
		return func() tea.Msg {
			
			if featureName == "" {
				fmt.Println("Error: feature name cannot be empty")
				return fmt.Errorf("feature name cannot be empty")
			}
			
			// Create the feature config
			feature := config.FeatureConfig{
				Feature:   featureName,
				Endpoints: []config.Endpoint{},
			}
			
			fmt.Printf("Creating feature: %+v\n", feature)
			
			// Create the feature using the mock manager
			if err := m.MockManager.CreateFeature(feature); err != nil {
				errMsg := fmt.Sprintf("Failed to create feature: %v", err)
				fmt.Println(errMsg)
				return fmt.Errorf(errMsg)
			}
			
			fmt.Println("Feature created successfully, initializing features list")
			
			// Update the features list
			m.initFeaturesList()
			
			// Select the new feature
			for i, item := range m.featuresList.Items() {
				if fi, ok := item.(featureItem); ok && fi.name == featureName {
					m.featuresList.Select(i)
					break
				}
			}
			
			m.selectedFeature = featureName
			m.updateEndpointsList()
			
			// Reload the server if it's running
			if m.Server.IsRunning() {
				if err := m.Server.Reload(); err != nil {
					fmt.Printf("Error reloading server: %v\n", err)
					return fmt.Errorf("failed to reload server: %v", err)
				}
			}
			
			fmt.Println("Feature creation completed successfully")
			
			// Return a custom message for smoother UI updates
			return customUpdateMsg{
				action: "feature_created",
				name:   featureName,
			}
		}
	}
	
	m.dialogCancelFn = func() tea.Cmd {
		return func() tea.Msg {
			fmt.Println("Feature creation cancelled")
			return nil
		}
	}
}

// showNewEndpointDialog shows the new endpoint dialog
func (m *Model) showNewEndpointDialog() {
	// Check if a feature is selected
	if m.selectedFeature == "" {
		return
	}
	
	// Clear any existing dialog state
	m.textInputs = nil
	m.dialogConfirmFn = nil
	m.dialogCancelFn = nil
	
	// Set dialog properties
	m.activeDialog = NewEndpointDialog
	m.dialogTitle = "Create New Endpoint"
	m.dialogContent = ""
	
	// Create text inputs with consistent width and styling
	idInput := textinput.New()
	idInput.Placeholder = "Endpoint ID"
	idInput.Focus()
	idInput.CharLimit = 32
	idInput.Width = 40
	
	methodInput := textinput.New()
	methodInput.Placeholder = "Method (GET, POST, PUT, DELETE)"
	methodInput.CharLimit = 10
	methodInput.Width = 40
	
	pathInput := textinput.New()
	pathInput.Placeholder = "Path (e.g., /api/users/:id)"
	pathInput.CharLimit = 100
	pathInput.Width = 40
	
	// Store the text inputs in the model
	m.textInputs = []textinput.Model{idInput, methodInput, pathInput}
	
	// Set the confirm function - this will be called when Enter is pressed
	m.dialogConfirmFn = func() tea.Cmd {
		// Capture the input values now, before text inputs are cleared
		var id, method, path string
		if len(m.textInputs) >= 3 {
			id = strings.TrimSpace(m.textInputs[0].Value())
			method = strings.TrimSpace(m.textInputs[1].Value())
			path = strings.TrimSpace(m.textInputs[2].Value())
		}
		
		return func() tea.Msg {
			// Debug print to console
			fmt.Printf("Creating new endpoint: %s %s %s\n", id, method, path)
			
			// Validate inputs
			if id == "" || method == "" || path == "" {
				fmt.Println("Error: all fields are required")
				return fmt.Errorf("all fields are required")
			}
			
			// Validate ID (alphanumeric and hyphens only)
			for _, c := range id {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
					fmt.Println("Error: endpoint ID can only contain letters, numbers, hyphens, and underscores")
					return fmt.Errorf("endpoint ID can only contain letters, numbers, hyphens, and underscores")
				}
			}
			
			// Validate method
			method = strings.ToUpper(method)
			if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" && method != "OPTIONS" && method != "HEAD" {
				fmt.Printf("Error: invalid HTTP method: %s\n", method)
				return fmt.Errorf("invalid HTTP method: %s", method)
			}
			
			// Validate path (must start with /)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
				fmt.Printf("Added leading slash to path: %s\n", path)
			}
			
			// Create a basic endpoint with a default response
			endpoint := config.Endpoint{
				ID:              id,
				Method:          method,
				Path:            path,
				Active:          true,
				DefaultResponse: "default",
				Responses: map[string]config.Response{
					"default": {
						Status: 200,
						Headers: map[string]string{
							"Content-Type": "application/json",
						},
						Body: map[string]interface{}{
							"message": "This is a default response",
						},
						Delay: 0,
					},
				},
			}
			
			fmt.Printf("Creating endpoint in feature '%s': %+v\n", m.selectedFeature, endpoint)
			
			// Create the endpoint using the mock manager
			if err := m.MockManager.CreateEndpoint(m.selectedFeature, endpoint); err != nil {
				errMsg := fmt.Sprintf("Failed to create endpoint: %v", err)
				fmt.Println(errMsg)
				return fmt.Errorf(errMsg)
			}
			
			fmt.Println("Endpoint created successfully, updating endpoints list")
			
			// Update the endpoints list
			m.updateEndpointsList()
			
			// Select the new endpoint
			for i, item := range m.endpointsList.Items() {
				if ei, ok := item.(endpointItem); ok && ei.id == id {
					m.endpointsList.Select(i)
					break
				}
			}
			
			// Reload the server if it's running
			if m.Server.IsRunning() {
				if err := m.Server.Reload(); err != nil {
					fmt.Printf("Error reloading server: %v\n", err)
					return fmt.Errorf("failed to reload server: %v", err)
				}
			}
			
			fmt.Println("Endpoint creation completed successfully")
			
			// Return a custom message for smoother UI updates
			return customUpdateMsg{
				action: "endpoint_created",
				name:   m.selectedFeature,
				id:     id,
			}
		}
	}
	
	m.dialogCancelFn = func() tea.Cmd {
		return func() tea.Msg {
			fmt.Println("Endpoint creation cancelled")
			return nil
		}
	}
}

// showDeleteConfirmDialog shows the delete confirmation dialog
func (m *Model) showDeleteConfirmDialog() {
	var item string
	var itemType string
	var confirmFn func() func() tea.Msg
	
	if m.activePanel == FeaturesPanel {
		if i, ok := m.featuresList.SelectedItem().(featureItem); ok {
			item = i.name
			itemType = "feature"
			confirmFn = func() func() tea.Msg {
				return m.deleteFeature
			}
		}
	} else {
		if i, ok := m.endpointsList.SelectedItem().(endpointItem); ok {
			item = i.id
			itemType = "endpoint"
			confirmFn = func() func() tea.Msg {
				return m.deleteEndpoint
			}
		}
	}
	
	if item == "" {
		// Nothing selected, don't show dialog
		return
	}
	
	// Clear any existing dialog state
	m.textInputs = nil
	m.dialogConfirmFn = nil
	m.dialogCancelFn = nil
	
	// Set dialog properties
	m.activeDialog = DeleteConfirmDialog
	m.dialogTitle = "Confirm Delete"
	m.dialogContent = fmt.Sprintf("Are you sure you want to delete this %s?\n\n%s", itemType, item)
	
	// Set the confirm function
	m.dialogConfirmFn = func() tea.Cmd {
		return func() tea.Msg {
			if confirmFn != nil {
				return confirmFn()()
			}
			return nil
		}
	}
	
	// Set the cancel function
	m.dialogCancelFn = func() tea.Cmd {
		return func() tea.Msg {
			fmt.Println("Delete operation cancelled")
			return nil
		}
	}
}

// showProxyConfigDialog shows the proxy configuration dialog
func (m *Model) showProxyConfigDialog() {
	// Clear any existing dialog state
	m.textInputs = nil
	m.dialogConfirmFn = nil
	m.dialogCancelFn = nil
	
	// Set dialog properties
	m.activeDialog = ProxyConfigDialog
	m.dialogTitle = "Proxy Configuration"
	m.dialogContent = ""
	
	// Create text input for proxy target with consistent styling
	targetInput := textinput.New()
	targetInput.Placeholder = "Proxy target URL (e.g., http://localhost:8080)"
	targetInput.Focus()
	targetInput.CharLimit = 100
	targetInput.Width = 50  // Slightly wider for URLs
	
	// Safely get the current proxy target URL
	currentTarget := m.ProxyManager.GetTargetURL()
	if currentTarget != "" {
		targetInput.SetValue(currentTarget)
	}
	
	m.textInputs = []textinput.Model{targetInput}
	
	m.dialogConfirmFn = func() tea.Cmd {
		return func() tea.Msg {
			// Safety check for text inputs
			if len(m.textInputs) == 0 {
				fmt.Println("Error: text inputs array is empty")
				return fmt.Errorf("text inputs array is empty")
			}
			
			return m.updateProxyConfig()()
		}
	}
	
	m.dialogCancelFn = func() tea.Cmd {
		return func() tea.Msg {
			fmt.Println("Proxy configuration cancelled")
			return nil
		}
	}
}

// createNewFeature creates a new feature
func (m *Model) createNewFeature() func() tea.Msg {
	return func() tea.Msg {
		// Safety check for text inputs
		if len(m.textInputs) == 0 {
			return fmt.Errorf("no text inputs available")
		}
		
		// Get the feature name from the text input
		featureName := strings.TrimSpace(m.textInputs[0].Value())
		
		if featureName == "" {
			fmt.Println("Error: feature name cannot be empty")
			return fmt.Errorf("feature name cannot be empty")
		}
		
		// Validate feature name (alphanumeric and hyphens only)
		for _, c := range featureName {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
				fmt.Println("Error: feature name can only contain letters, numbers, hyphens, and underscores")
				return fmt.Errorf("feature name can only contain letters, numbers, hyphens, and underscores")
			}
		}
		
		// Create the feature config
		feature := config.FeatureConfig{
			Feature:   featureName,
			Endpoints: []config.Endpoint{},
		}
		
		fmt.Printf("Creating feature: %+v\n", feature)
		
		// Create the feature using the mock manager
		if err := m.MockManager.CreateFeature(feature); err != nil {
			errMsg := fmt.Sprintf("Failed to create feature: %v", err)
			fmt.Println(errMsg)
			return fmt.Errorf(errMsg)
		}
		
		fmt.Println("Feature created successfully, initializing features list")
		
		// Update the features list
		m.initFeaturesList()
		
		// Select the new feature
		for i, item := range m.featuresList.Items() {
			if fi, ok := item.(featureItem); ok && fi.name == featureName {
				m.featuresList.Select(i)
				break
			}
		}
		
		m.selectedFeature = featureName
		m.updateEndpointsList()
		
		// Reload the server if it's running
		if m.Server.IsRunning() {
			if err := m.Server.Reload(); err != nil {
				fmt.Printf("Error reloading server: %v\n", err)
				return fmt.Errorf("failed to reload server: %v", err)
			}
		}
		
		fmt.Println("Feature creation completed successfully")
		return nil
	}
}

// createNewEndpoint creates a new endpoint
func (m *Model) createNewEndpoint() func() tea.Msg {
	return func() tea.Msg {
		// Safety check for text inputs
		if len(m.textInputs) < 3 {
			return fmt.Errorf("not enough text inputs available")
		}
		
		// Get values from text inputs
		id := strings.TrimSpace(m.textInputs[0].Value())
		method := strings.TrimSpace(m.textInputs[1].Value())
		path := strings.TrimSpace(m.textInputs[2].Value())
		
		// Validate inputs
		if id == "" || method == "" || path == "" {
			fmt.Println("Error: all fields are required")
			return fmt.Errorf("all fields are required")
		}
		
		// Validate ID (alphanumeric and hyphens only)
		for _, c := range id {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
				fmt.Println("Error: endpoint ID can only contain letters, numbers, hyphens, and underscores")
				return fmt.Errorf("endpoint ID can only contain letters, numbers, hyphens, and underscores")
			}
		}
		
		// Validate method
		method = strings.ToUpper(method)
		if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" && method != "OPTIONS" && method != "HEAD" {
			fmt.Printf("Error: invalid HTTP method: %s\n", method)
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
		
		// Validate path (must start with /)
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
			fmt.Printf("Added leading slash to path: %s\n", path)
		}
		
		// Create a basic endpoint with a default response
		endpoint := config.Endpoint{
			ID:              id,
			Method:          method,
			Path:            path,
			Active:          true,
			DefaultResponse: "default",
			Responses: map[string]config.Response{
				"default": {
					Status: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"message": "This is a default response",
					},
					Delay: 0,
				},
			},
		}
		
		fmt.Printf("Creating endpoint in feature '%s': %+v\n", m.selectedFeature, endpoint)
		
		// Create the endpoint using the mock manager
		if err := m.MockManager.CreateEndpoint(m.selectedFeature, endpoint); err != nil {
			errMsg := fmt.Sprintf("Failed to create endpoint: %v", err)
			fmt.Println(errMsg)
			return fmt.Errorf(errMsg)
		}
		
		fmt.Println("Endpoint created successfully, updating endpoints list")
		
		// Update the endpoints list
		m.updateEndpointsList()
		
		// Select the new endpoint
		for i, item := range m.endpointsList.Items() {
			if ei, ok := item.(endpointItem); ok && ei.id == id {
				m.endpointsList.Select(i)
				break
			}
		}
		
		// Reload the server if it's running
		if m.Server.IsRunning() {
			if err := m.Server.Reload(); err != nil {
				return fmt.Errorf("failed to reload server: %v", err)
			}
		}
		return nil
	}
}

// deleteFeature deletes the selected feature
func (m *Model) deleteFeature() tea.Msg {
	item, ok := m.featuresList.SelectedItem().(featureItem)
	if !ok {
		return fmt.Errorf("no feature selected")
	}
	
	if err := m.MockManager.DeleteFeature(item.name); err != nil {
		return fmt.Errorf("failed to delete feature: %w", err)
	}
	m.initFeaturesList()
	
	// Select the first feature if available
	if len(m.featuresList.Items()) > 0 {
		m.featuresList.Select(0)
		if i, ok := m.featuresList.SelectedItem().(featureItem); ok {
			m.selectedFeature = i.name
		}
	} else {
		m.selectedFeature = ""
	}
	
	m.updateEndpointsList()
	
	if m.Server.IsRunning() {
		if err := m.Server.Reload(); err != nil {
			return fmt.Errorf("failed to reload server: %v", err)
		}
	}
	
	// Return a custom message for smoother UI updates
	return customUpdateMsg{
		action: "feature_deleted",
		name:   item.name,
	}
}

// deleteEndpoint deletes the selected endpoint
func (m *Model) deleteEndpoint() tea.Msg {
	item, ok := m.endpointsList.SelectedItem().(endpointItem)
	if !ok {
		return fmt.Errorf("no endpoint selected")
	}
	
	if err := m.MockManager.DeleteEndpoint(m.selectedFeature, item.id); err != nil {
		return fmt.Errorf("failed to delete endpoint: %w", err)
	}
	m.updateEndpointsList()
	
	if m.Server.IsRunning() {
		if err := m.Server.Reload(); err != nil {
			return fmt.Errorf("failed to reload server: %v", err)
		}
	}
	
	// Return a custom message for smoother UI updates
	return customUpdateMsg{
		action: "endpoint_deleted",
		name:   m.selectedFeature,
		id:     item.id,
	}
}

// updateProxyConfig updates the proxy configuration
func (m *Model) updateProxyConfig() func() tea.Msg {
	return func() tea.Msg {
		// Safety check for text inputs
		if len(m.textInputs) == 0 {
			return fmt.Errorf("no text inputs available")
		}
		
		// Get the target from the text input
		target := strings.TrimSpace(m.textInputs[0].Value())
		
		if target == "" {
			return fmt.Errorf("proxy target cannot be empty")
		}
		
		// Validate URL format
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
			return fmt.Errorf("proxy target must start with http:// or https://")
		}
		
		// Update the proxy target
		if err := m.ProxyManager.UpdateTarget(target); err != nil {
			return fmt.Errorf("failed to update proxy target: %w", err)
		}
		
		return nil
	}
}