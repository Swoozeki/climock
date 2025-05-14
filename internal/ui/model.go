package ui

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"kohofinancial/mockoho/internal/config"
	"kohofinancial/mockoho/internal/logger"
	"kohofinancial/mockoho/internal/mock"
	"kohofinancial/mockoho/internal/proxy"
	"kohofinancial/mockoho/internal/server"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Model represents the UI model
type Model struct {
	Config       *config.Config
	MockManager  *mock.Manager
	ProxyManager *proxy.Manager
	Server       *server.Server
	
	// UI state
	activePanel     Panel
	featuresList    list.Model
	endpointsList   list.Model
	selectedFeature string
	width           int
	height          int
	keyMap          KeyMap
	help            help.Model
	
	// Dialog state
	activeDialog    DialogType
	textInputs      []textinput.Model
	dialogTitle     string
	dialogContent   string
	dialogConfirmFn func() tea.Cmd
	dialogCancelFn  func() tea.Cmd
	
	// Performance optimization
	lastUpdate time.Time
	styles     struct {
		header         lipgloss.Style
		featureTitle   lipgloss.Style
		endpointsTitle lipgloss.Style
		features       lipgloss.Style
		endpoints      lipgloss.Style
		footer         lipgloss.Style
	}
}

// customUpdateMsg is a custom message type for smoother UI updates
type customUpdateMsg struct {
	action string
	name   string
	id     string
	// Additional fields for more specific updates
	feature  string
	active   bool
	response string
}

// New creates a new UI model
func New(cfg *config.Config, mockManager *mock.Manager, proxyManager *proxy.Manager, srv *server.Server) *Model {
	keyMap := DefaultKeyMap()
	helpModel := help.New()
	helpModel.ShowAll = false

	m := &Model{
		Config:       cfg,
		MockManager:  mockManager,
		ProxyManager: proxyManager,
		Server:       srv,
		activePanel:  FeaturesPanel,
		keyMap:       keyMap,
		help:         helpModel,
		// Set initial dimensions to reasonable defaults
		width:        100,
		height:       30,
		// Initialize dialog state
		activeDialog:  NoDialog,
		textInputs:    nil,
		dialogTitle:   "",
		dialogContent: "",
		dialogConfirmFn: nil,
		dialogCancelFn:  nil,
		// Initialize performance optimization
		lastUpdate: time.Now(),
	}

	// Initialize cached styles
	m.initStyles()
	
	// Initialize feature list
	m.initFeaturesList()
	
	// Initialize endpoints list
	m.initEndpointsList()
	
	// Set initial list dimensions
	m.featuresList.SetSize(m.width/4, m.height-6)
	m.endpointsList.SetSize(3*m.width/4, m.height-6)
	m.help.Width = m.width

	return m
}

// initStyles initializes cached styles for better performance
func (m *Model) initStyles() {
	// Header style - removed bottom border
	m.styles.header = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		BorderBottom(false). // No bottom border
		Padding(1, 2)
	
	// Panel title styles
	m.styles.featureTitle = lipgloss.NewStyle().
		Width(m.width/4).
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		BorderBottom(true).
		Padding(0, 1)
	
	m.styles.endpointsTitle = lipgloss.NewStyle().
		Width(3*m.width/4).
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		BorderBottom(true).
		Padding(0, 1)
	
	// List styles
	m.styles.features = lipgloss.NewStyle().
		Width(m.width/4).
		Padding(0, 1)
	
	m.styles.endpoints = lipgloss.NewStyle().
		Width(3*m.width/4).
		Padding(0, 1)
	
	// Footer style - removed top border
	m.styles.footer = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		BorderTop(false). // No top border
		Padding(0, 2)
}

// Init initializes the UI model
func (m *Model) Init() tea.Cmd {
	// Initialize list delegates based on active panel
	m.updateListDelegatesForActivePanel()
	
	// Return commands to initialize the terminal and UI
	return tea.Batch(
		// Enter alt screen without clearing first (reduces flicker)
		tea.EnterAltScreen,
		
		// Get the terminal size more gently
		func() tea.Msg {
			// Get the current terminal size
			width, height, _ := term.GetSize(int(os.Stdout.Fd()))
			
			// Only update if dimensions have changed
			if width != m.width || height != m.height {
				m.width = width
				m.height = height
				
				// Update list dimensions
				topHeight := 4 // Header height
				bottomHeight := 2 // Footer height
				listHeight := height - topHeight - bottomHeight
				
				// Adjust widths to account for borders
				featureWidth := width/4 - 2
				endpointWidth := 3*width/4 - 2
				
				m.featuresList.SetSize(featureWidth, listHeight)
				m.endpointsList.SetSize(endpointWidth, listHeight)
				m.help.Width = width
			}
			
			// Return a window size message with no logging
			return tea.WindowSizeMsg{
				Width:  width,
				Height: height,
			}
		},
	)
}

// createCompactDelegate creates a compact list delegate with optimized styles
func (m *Model) createCompactDelegate(showDescription bool) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	
	// Create a custom style for the delegate
	styles := delegate.Styles
	normalTitle := styles.NormalTitle.UnsetMargins().PaddingTop(0).PaddingBottom(0)
	styles.NormalTitle = normalTitle
	
	normalDesc := styles.NormalDesc.UnsetMargins().PaddingTop(0).PaddingBottom(0)
	styles.NormalDesc = normalDesc
	
	selectedTitle := styles.SelectedTitle.UnsetMargins().PaddingTop(0).PaddingBottom(0)
	styles.SelectedTitle = selectedTitle
	
	selectedDesc := styles.SelectedDesc.UnsetMargins().PaddingTop(0).PaddingBottom(0)
	styles.SelectedDesc = selectedDesc
	
	// Create a new delegate with the custom styles
	compactDelegate := list.NewDefaultDelegate()
	compactDelegate.Styles = styles
	compactDelegate.ShowDescription = showDescription
	
	return compactDelegate
}
// initFeaturesList initializes the features list
func (m *Model) initFeaturesList() {
	// Create a compact delegate with no description
	compactDelegate := m.createCompactDelegate(false)
	
	items := []list.Item{}
	
	// Add features from config
	for feature := range m.Config.Mocks {
		items = append(items, featureItem{name: feature})
	}
	
	// Create the list with proper dimensions
	listHeight := m.height - 6 // Account for header and footer
	if listHeight < 1 {
		listHeight = 10 // Default if height not set yet
	}
	
	// Adjust width to account for borders
	featureWidth := m.width/4 - 2
	
	m.featuresList = list.New(items, compactDelegate, featureWidth, listHeight)
	m.featuresList.Title = "Features"
	m.featuresList.SetShowStatusBar(false)
	m.featuresList.SetFilteringEnabled(false)
	m.featuresList.SetShowHelp(false)
	
	// Update delegates based on active panel
	m.updateListDelegatesForActivePanel()
	
	// Select the first feature if available
	if len(items) > 0 {
		m.featuresList.Select(0) // Explicitly select the first item
		if fi, ok := items[0].(featureItem); ok {
			m.selectedFeature = fi.name
		}
	} else {
		// No features available
		m.selectedFeature = ""
	}
}

// createEndpointItems creates endpoint items for the list
func (m *Model) createEndpointItems() []list.Item {
	items := []list.Item{}
	
	// Add endpoints from selected feature
	if m.selectedFeature != "" {
		if featureConfig, ok := m.Config.Mocks[m.selectedFeature]; ok {
			for _, endpoint := range featureConfig.Endpoints {
				// Get all response names and sort them alphabetically for consistent order
				var allResponses []string
				for name := range endpoint.Responses {
					allResponses = append(allResponses, name)
				}
				
				// Sort responses alphabetically using Go's built-in sort package
				sort.Strings(allResponses)
				
				items = append(items, endpointItem{
					id:              endpoint.ID,
					method:          endpoint.Method,
					path:            endpoint.Path,
					active:          endpoint.Active,
					defaultResponse: endpoint.DefaultResponse,
					responses:       allResponses,
				})
			}
		}
	}
	
	return items
}

// initEndpointsList initializes the endpoints list
func (m *Model) initEndpointsList() {
	// Create a compact delegate with description
	compactDelegate := m.createCompactDelegate(true)
	
	// Get endpoint items
	items := m.createEndpointItems()
	
	// Create the list with proper dimensions
	listHeight := m.height - 6 // Account for header and footer
	if listHeight < 1 {
		listHeight = 10 // Default if height not set yet
	}
	
	// Adjust width to account for borders
	endpointWidth := 3*m.width/4 - 2
	
	m.endpointsList = list.New(items, compactDelegate, endpointWidth, listHeight)
	m.endpointsList.Title = fmt.Sprintf("Endpoints (%s)", m.selectedFeature)
	m.endpointsList.SetShowStatusBar(false)
	m.endpointsList.SetFilteringEnabled(false)
	m.endpointsList.SetShowHelp(false)
}

// updateEndpointsList updates the endpoints list based on the selected feature
func (m *Model) updateEndpointsList() {
	// Save current selection index
	currentIndex := m.endpointsList.Index()
	
	// Get endpoint items using the shared function
	items := m.createEndpointItems()
	
	// Update just the items, not the entire list
	m.endpointsList.SetItems(items)
	m.endpointsList.Title = fmt.Sprintf("Endpoints (%s)", m.selectedFeature)
	
	// Restore selection if possible
	if currentIndex < len(m.endpointsList.Items()) {
		m.endpointsList.Select(currentIndex)
	}
	
	// Update delegates based on active panel
	m.updateListDelegatesForActivePanel()
}

// updateListDelegatesForActivePanel updates the list delegates based on the active panel
func (m *Model) updateListDelegatesForActivePanel() {
	// Create compact style function to reduce duplication
	makeCompactStyle := func(style lipgloss.Style) lipgloss.Style {
		return style.UnsetMargins().PaddingTop(0).PaddingBottom(0)
	}

	// Create delegates with appropriate styles based on active panel
	featuresDelegate := list.NewDefaultDelegate()
	endpointsDelegate := list.NewDefaultDelegate()
	
	// Make both delegates compact
	featuresDelegate.Styles.NormalTitle = makeCompactStyle(featuresDelegate.Styles.NormalTitle)
	featuresDelegate.Styles.NormalDesc = makeCompactStyle(featuresDelegate.Styles.NormalDesc)
	endpointsDelegate.Styles.NormalTitle = makeCompactStyle(endpointsDelegate.Styles.NormalTitle)
	endpointsDelegate.Styles.NormalDesc = makeCompactStyle(endpointsDelegate.Styles.NormalDesc)
	
	// Set selection styles based on active panel
	if m.activePanel == FeaturesPanel {
		// Keep features selection visible
		featuresDelegate.Styles.SelectedTitle = makeCompactStyle(featuresDelegate.Styles.SelectedTitle)
		featuresDelegate.Styles.SelectedDesc = makeCompactStyle(featuresDelegate.Styles.SelectedDesc)
		
		// Make endpoints selection less visible (same as normal)
		endpointsDelegate.Styles.SelectedTitle = endpointsDelegate.Styles.NormalTitle
		endpointsDelegate.Styles.SelectedDesc = endpointsDelegate.Styles.NormalDesc
	} else {
		// Make features selection less visible (same as normal)
		featuresDelegate.Styles.SelectedTitle = featuresDelegate.Styles.NormalTitle
		featuresDelegate.Styles.SelectedDesc = featuresDelegate.Styles.NormalDesc
		
		// Keep endpoints selection visible
		endpointsDelegate.Styles.SelectedTitle = makeCompactStyle(endpointsDelegate.Styles.SelectedTitle)
		endpointsDelegate.Styles.SelectedDesc = makeCompactStyle(endpointsDelegate.Styles.SelectedDesc)
	}
	
	// Hide description for features to save space
	featuresDelegate.ShowDescription = false
	
	// Update the lists with the new delegates
	m.featuresList.SetDelegate(featuresDelegate)
	m.endpointsList.SetDelegate(endpointsDelegate)
}

// Update updates the UI model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Throttle updates to max 30fps (about 33ms between updates)
	now := time.Now()
	if now.Sub(m.lastUpdate) < 33*time.Millisecond {
		// Skip non-essential updates if they come too quickly
		switch msg.(type) {
		case tea.WindowSizeMsg, tea.KeyMsg:
			// Always process these immediately
		default:
			// Delay other updates
			return m, tea.Tick(33*time.Millisecond-now.Sub(m.lastUpdate), func(t time.Time) tea.Msg {
				return msg
			})
		}
	}
	m.lastUpdate = now

	switch msg := msg.(type) {
	case customUpdateMsg:
		// Handle custom update messages for smoother UI updates
		switch msg.action {
		case "feature_created":
			// Feature was created, no need to force a full redraw
			// The lists have already been updated in the dialog confirm function
			
		case "endpoint_created":
			// Endpoint was created, no need to force a full redraw
			// The lists have already been updated in the dialog confirm function
			
		case "feature_deleted":
			// Feature was deleted, no need to force a full redraw
			// The lists have already been updated in the dialog confirm function
			
		case "endpoint_deleted":
			// Endpoint was deleted, no need to force a full redraw
			// The lists have already been updated in the dialog confirm function
			
		case "server_toggled":
			// Server was started or stopped, force a UI update
			// No additional action needed as the message itself triggers the update
			
		case "endpoint_updated":
			// Update just the specific endpoint in the list
			if msg.id != "" {
				for i, item := range m.endpointsList.Items() {
					if ei, ok := item.(endpointItem); ok && ei.id == msg.id {
						// Update just this item
						items := m.endpointsList.Items()
						endpoint, _ := m.Config.GetEndpoint(m.selectedFeature, msg.id)
						if endpoint != nil {
							// Get all response names
							var allResponses []string
							for name := range endpoint.Responses {
								allResponses = append(allResponses, name)
							}
							
							// Sort responses alphabetically using Go's built-in sort package
							sort.Strings(allResponses)
							
							items[i] = endpointItem{
								id:              endpoint.ID,
								method:          endpoint.Method,
								path:            endpoint.Path,
								active:          endpoint.Active,
								defaultResponse: endpoint.DefaultResponse,
								responses:       allResponses,
							}
							m.endpointsList.SetItems(items)
						}
						break
					}
				}
			}
		}
		
	case tea.WindowSizeMsg:
		// Handle window size changes
		m.width = msg.Width
		m.height = msg.Height
		
		// Update list dimensions
		topHeight := 4 // Header height
		bottomHeight := 2 // Footer height
		listHeight := m.height - topHeight - bottomHeight
		
		// Adjust widths to account for borders (subtract 2 for borders)
		featureWidth := m.width/4 - 2
		endpointWidth := 3*m.width/4 - 2
		
		m.featuresList.SetSize(featureWidth, listHeight)
		m.endpointsList.SetSize(endpointWidth, listHeight)
		
		m.help.Width = m.width
		
		// Update cached styles with new dimensions
		m.initStyles()
		
	case tea.KeyMsg:
		// Handle dialog-specific key presses
		if m.activeDialog != NoDialog {
			return m.updateDialog(msg)
		}

		// Handle global key presses
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keyMap.Left):
			if m.activePanel == EndpointsPanel {
				m.activePanel = FeaturesPanel
				m.updateListDelegatesForActivePanel()
			}
		case key.Matches(msg, m.keyMap.Right):
			if m.activePanel == FeaturesPanel {
				m.activePanel = EndpointsPanel
				m.updateListDelegatesForActivePanel()
			}
		case key.Matches(msg, m.keyMap.Help):
			m.activeDialog = HelpDialog
			// Initialize dialog content
			m.dialogTitle = "Mockoho Help"
			m.dialogContent = ""
			return m, nil
		case key.Matches(msg, m.keyMap.New):
			if m.activePanel == FeaturesPanel {
				// Always allow creating new features in the features panel
				m.showNewFeatureDialog()
			} else if m.activePanel == EndpointsPanel && m.selectedFeature != "" {
				// Only allow creating new endpoints if a feature is selected
				m.showNewEndpointDialog()
			}
			return m, nil
		case key.Matches(msg, m.keyMap.Delete):
			// Only show delete dialog if there's something to delete
			hasSelection := false
			if m.activePanel == FeaturesPanel {
				hasSelection = len(m.featuresList.Items()) > 0
			} else { // EndpointsPanel
				hasSelection = m.selectedFeature != "" && len(m.endpointsList.Items()) > 0
			}
			
			if hasSelection {
				m.showDeleteConfirmDialog()
			}
			return m, nil
		case key.Matches(msg, m.keyMap.Proxy):
			m.showProxyConfigDialog()
			return m, nil
		case key.Matches(msg, m.keyMap.Server):
			return m, m.toggleServer()
		case key.Matches(msg, m.keyMap.Reload):
			return m, m.reloadConfig
		case key.Matches(msg, m.keyMap.Toggle):
			// Only toggle if we're in the endpoints panel and there are endpoints
			if m.activePanel == EndpointsPanel && m.selectedFeature != "" && len(m.endpointsList.Items()) > 0 {
				return m, m.toggleEndpoint()
			}
		case key.Matches(msg, m.keyMap.Response):
			// Only cycle response if we're in the endpoints panel and there are endpoints
			if m.activePanel == EndpointsPanel && m.selectedFeature != "" && len(m.endpointsList.Items()) > 0 {
				return m, m.cycleResponse()
			}
		case key.Matches(msg, m.keyMap.Open):
			// Only try to open if there's something to open
			hasSelection := false
			if m.activePanel == FeaturesPanel {
				hasSelection = len(m.featuresList.Items()) > 0
			} else { // EndpointsPanel
				hasSelection = m.selectedFeature != "" && len(m.endpointsList.Items()) > 0
			}
			
			if hasSelection {
				return m, m.openInEditor
			}
			return m, nil
		}
	}

	// Update the active list
	if m.activePanel == FeaturesPanel {
		var listCmd tea.Cmd
		m.featuresList, listCmd = m.featuresList.Update(msg)
		cmds = append(cmds, listCmd)
		
		// Update selected feature when list selection changes
		if i, ok := m.featuresList.SelectedItem().(featureItem); ok {
			if m.selectedFeature != i.name {
				m.selectedFeature = i.name
				m.updateEndpointsList()
			}
		}
	} else {
		var listCmd tea.Cmd
		m.endpointsList, listCmd = m.endpointsList.Update(msg)
		cmds = append(cmds, listCmd)
	}

	return m, tea.Batch(cmds...)
}

// toggleServer toggles the server on/off
func (m *Model) toggleServer() tea.Cmd {
	return func() tea.Msg {
		if m.Server.IsRunning() {
			logger.Info("User requested to stop server")
			if err := m.Server.Stop(); err != nil {
				logger.Error("Failed to stop server: %v", err)
				return fmt.Errorf("failed to stop server: %v", err)
			}
			// Return a custom update message to trigger UI refresh
			return customUpdateMsg{action: "server_toggled", active: false}
		} else {
			logger.Info("User requested to start server")
			if err := m.Server.Start(); err != nil {
				logger.Error("Failed to start server: %v", err)
				return fmt.Errorf("failed to start server: %v", err)
			}
			// Return a custom update message to trigger UI refresh
			return customUpdateMsg{action: "server_toggled", active: true}
		}
	}
}

// reloadConfig reloads the configuration
func (m *Model) reloadConfig() tea.Msg {
	if err := m.Config.Load(); err != nil {
		return err
	}
	
	m.initFeaturesList()
	m.updateEndpointsList()
	
	if m.Server.IsRunning() {
		if err := m.Server.Reload(); err != nil {
			return err
		}
	}
	
	return nil
}

// toggleEndpoint toggles the selected endpoint
func (m *Model) toggleEndpoint() tea.Cmd {
	return func() tea.Msg {
		// Check if we're in the endpoints panel and have endpoints
		if m.activePanel != EndpointsPanel || m.selectedFeature == "" || len(m.endpointsList.Items()) == 0 {
			return nil
		}
		
		item, ok := m.endpointsList.SelectedItem().(endpointItem)
		if !ok {
			return nil
		}
		
		if err := m.MockManager.ToggleEndpoint(m.selectedFeature, item.id); err != nil {
			return err
		}
		
		if m.Server.IsRunning() {
			if err := m.Server.Reload(); err != nil {
				return err
			}
		}
		
		// Return a custom update message instead of forcing a full redraw
		return customUpdateMsg{
			action:  "endpoint_updated",
			id:      item.id,
			feature: m.selectedFeature,
		}
	}
}

// cycleResponse cycles through the available responses for the selected endpoint
func (m *Model) cycleResponse() tea.Cmd {
	return func() tea.Msg {
		// Check if we're in the endpoints panel and have endpoints
		if m.activePanel != EndpointsPanel || m.selectedFeature == "" || len(m.endpointsList.Items()) == 0 {
			return nil
		}
		
		item, ok := m.endpointsList.SelectedItem().(endpointItem)
		if !ok {
			return nil
		}
		
		endpoint, err := m.Config.GetEndpoint(m.selectedFeature, item.id)
		if err != nil {
			return err
		}
		
		// Get all response names
		var responses []string
		for name := range endpoint.Responses {
			responses = append(responses, name)
		}
		
		if len(responses) == 0 {
			return nil
		}
		
		// Find the current default response
		currentIndex := -1
		for i, name := range responses {
			if name == endpoint.DefaultResponse {
				currentIndex = i
				break
			}
		}
		
		// Move to the next response linearly
		nextIndex := currentIndex + 1
		// If we're at the end, go back to the first response
		if nextIndex >= len(responses) {
			nextIndex = 0
		}
		nextResponse := responses[nextIndex]
		
		if err := m.MockManager.SetDefaultResponse(m.selectedFeature, item.id, nextResponse); err != nil {
			return err
		}
		
		if m.Server.IsRunning() {
			if err := m.Server.Reload(); err != nil {
				return err
			}
		}
		
		// Return a custom update message instead of forcing a full redraw
		return customUpdateMsg{
			action:   "endpoint_updated",
			id:       item.id,
			feature:  m.selectedFeature,
			response: nextResponse,
		}
	}
}

// openInEditor opens the selected feature or endpoint in the editor
func (m *Model) openInEditor() tea.Msg {
	var filePath string
	var line int
	
	// Check if there are items to select from
	if m.activePanel == FeaturesPanel {
		if len(m.featuresList.Items()) == 0 {
			return nil // No features available, silently do nothing
		}
		
		item, ok := m.featuresList.SelectedItem().(featureItem)
		if !ok {
			return fmt.Errorf("no feature selected")
		}
		
		filePath = fmt.Sprintf("%s/%s.json", m.Config.BaseDir, item.name)
		line = 1
	} else {
		if m.selectedFeature == "" || len(m.endpointsList.Items()) == 0 {
			return nil // No endpoints available, silently do nothing
		}
		
		endpoint, ok := m.endpointsList.SelectedItem().(endpointItem)
		if !ok {
			return fmt.Errorf("no endpoint selected")
		}
		
		filePath = fmt.Sprintf("%s/%s.json", m.Config.BaseDir, m.selectedFeature)
		
		// Find the actual line number of the endpoint in the file
		line = findEndpointLineNumber(filePath, endpoint.id)
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	
	// Get editor command and args
	command := m.Config.Global.Editor.Command
	if command == "" {
		return fmt.Errorf("editor command not configured")
	}
	
	// Create a new slice for args to avoid modifying the original
	args := make([]string, 0, len(m.Config.Global.Editor.Args))
	
	// Replace placeholders in args
	for _, arg := range m.Config.Global.Editor.Args {
		newArg := strings.ReplaceAll(arg, "{file}", filePath)
		newArg = strings.ReplaceAll(newArg, "{line}", fmt.Sprintf("%d", line))
		args = append(args, newArg)
	}
	
	// Execute the editor command
	
	// Execute the editor command
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start editor: %v", err)
	}
	
	// Don't wait for the editor to close
	return nil
}

// findEndpointLineNumber finds the line number of an endpoint in a JSON file
func findEndpointLineNumber(filePath, endpointID string) int {
	// Default line number if we can't find the exact position
	defaultLine := 1
	
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return defaultLine
	}
	
	// Convert to string for line-by-line processing
	content := string(data)
	lines := strings.Split(content, "\n")
	
	// First, find the endpoints array
	endpointsStartLine := -1
	for i, line := range lines {
		if strings.Contains(line, `"endpoints":`) {
			endpointsStartLine = i
			break
		}
	}
	
	if endpointsStartLine == -1 {
		return defaultLine
	}
	
	// Now search for the endpoint with the matching ID
	inEndpoint := false
	endpointStartLine := -1
	idLine := -1
	pathLine := -1
	
	for i := endpointsStartLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		
		// Start of an endpoint object
		if line == "{" && !inEndpoint {
			inEndpoint = true
			endpointStartLine = i
			continue
		}
		
		// End of an endpoint object
		if line == "}" && inEndpoint {
			// If we found the ID but not the path, reset and continue
			if idLine > 0 && pathLine == -1 {
				inEndpoint = false
				endpointStartLine = -1
				idLine = -1
				continue
			}
			
			// If we found both ID and path, we're done
			if idLine > 0 && pathLine > 0 {
				return pathLine + 1 // Return the path line (1-based)
			}
		}
		
		// Look for the ID field
		if inEndpoint && strings.Contains(line, `"id":`) && strings.Contains(line, `"`+endpointID+`"`) {
			idLine = i
		}
		
		// Look for the path field if we've already found the ID
		if inEndpoint && idLine > 0 && strings.Contains(line, `"path":`) {
			pathLine = i
		}
	}
	
	// If we found the ID but not the path, return the ID line
	if idLine > 0 {
		return idLine + 1
	}
	
	// If we found the endpoint start but not the ID or path, return the endpoint start line
	if endpointStartLine > 0 {
		return endpointStartLine + 1
	}
	
	return defaultLine
}

// updateDialog updates the active dialog
func (m *Model) updateDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Safety check - if somehow we get here with NoDialog, return to normal UI
	if m.activeDialog == NoDialog {
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		// Cancel the dialog
		m.activeDialog = NoDialog
		
		// Store the cancel function before clearing dialog state
		var cancelFn func() tea.Cmd
		if m.dialogCancelFn != nil {
			cancelFn = m.dialogCancelFn
		}
		
		// Clear dialog state
		m.textInputs = nil
		m.dialogTitle = ""
		m.dialogContent = ""
		m.dialogCancelFn = nil
		m.dialogConfirmFn = nil
		
		// Execute cancel function if available
		if cancelFn != nil {
			return m, cancelFn()
		}
		
		return m, nil
		
	case tea.KeyEnter:
		// Confirm the dialog
		if m.activeDialog == HelpDialog {
			m.activeDialog = NoDialog
			m.dialogTitle = ""
			m.dialogContent = ""
			return m, nil
		}
		
		// Execute the confirm function if available
		if m.dialogConfirmFn != nil {
			// Store the confirm function before clearing dialog state
			confirmFn := m.dialogConfirmFn
			
			// Execute the confirm function BEFORE clearing any state
			// This ensures the text inputs are still available when the command is executed
			cmd := confirmFn()
			
			// Now clear dialog state
			m.activeDialog = NoDialog
			m.dialogTitle = ""
			m.dialogContent = ""
			m.dialogConfirmFn = nil
			m.dialogCancelFn = nil
			m.textInputs = nil
			
			return m, cmd
		}
		
		// If no confirm function, just close the dialog
		m.activeDialog = NoDialog
		m.textInputs = nil
		m.dialogTitle = ""
		m.dialogContent = ""
		m.dialogConfirmFn = nil
		m.dialogCancelFn = nil
		return m, nil
		
	case tea.KeyTab:
		// Handle tab navigation between text inputs
		if len(m.textInputs) > 1 {
			// Find the currently focused input
			focusedIndex := -1
			for i, ti := range m.textInputs {
				if ti.Focused() {
					focusedIndex = i
					break
				}
			}
			
			// If no input is focused, focus the first one
			if focusedIndex == -1 {
				m.textInputs[0].Focus()
				return m, nil
			}
			
			// Blur the current input
			m.textInputs[focusedIndex].Blur()
			
			// Focus the next input (or wrap around to the first)
			nextIndex := (focusedIndex + 1) % len(m.textInputs)
			m.textInputs[nextIndex].Focus()
			// Focus moved to next input
			
			
			return m, nil
		}
		
	case tea.KeyShiftTab:
		// Handle shift+tab navigation between text inputs (backwards)
		if len(m.textInputs) > 1 {
			// Find the currently focused input
			focusedIndex := -1
			for i, ti := range m.textInputs {
				if ti.Focused() {
					focusedIndex = i
					break
				}
			}
			
			// If no input is focused, focus the last one
			if focusedIndex == -1 {
				lastIndex := len(m.textInputs) - 1
				m.textInputs[lastIndex].Focus()
				return m, nil
			}
			
			// Blur the current input
			m.textInputs[focusedIndex].Blur()
			
			// Focus the previous input (or wrap around to the last)
			prevIndex := (focusedIndex - 1 + len(m.textInputs)) % len(m.textInputs)
			m.textInputs[prevIndex].Focus()
			
			// Focus moved to previous input
			
			return m, nil
		}
		
	default:
		// Update text inputs if any
		if len(m.textInputs) > 0 {
			// Create a slice to hold commands
			cmds := make([]tea.Cmd, len(m.textInputs))
			
			// Update each text input
			for i := range m.textInputs {
				m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
			}
			
			// Ensure at least one input is focused
			focusedFound := false
			for _, ti := range m.textInputs {
				if ti.Focused() {
					focusedFound = true
					break
				}
			}
			
			// If no input is focused, focus the first one
			if !focusedFound && len(m.textInputs) > 0 {
				m.textInputs[0].Focus()
			}
			
			return m, tea.Batch(cmds...)
		} else if m.activeDialog == HelpDialog {
			// Any key dismisses help dialog
			m.activeDialog = NoDialog
			return m, nil
		}
	}
	
	return m, nil
}