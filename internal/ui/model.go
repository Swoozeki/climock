package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockoho/mockoho/internal/config"
	"github.com/mockoho/mockoho/internal/mock"
	"github.com/mockoho/mockoho/internal/proxy"
	"github.com/mockoho/mockoho/internal/server"
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
	endpoint string
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
	// Header style
	m.styles.header = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)
	
	// Panel title styles
	m.styles.featureTitle = lipgloss.NewStyle().
		Width(m.width/4).
		Align(lipgloss.Left).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)
	
	m.styles.endpointsTitle = lipgloss.NewStyle().
		Width(3*m.width/4).
		Align(lipgloss.Left).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)
	
	// List styles
	m.styles.features = lipgloss.NewStyle().
		Width(m.width/4)
	
	m.styles.endpoints = lipgloss.NewStyle().
		Width(3*m.width/4)
	
	// Footer style
	m.styles.footer = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true)
}

// Init initializes the UI model
func (m *Model) Init() tea.Cmd {
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
				
				m.featuresList.SetSize(width/4, listHeight)
				m.endpointsList.SetSize(3*width/4, listHeight)
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

// initFeaturesList initializes the features list
func (m *Model) initFeaturesList() {
	delegate := list.NewDefaultDelegate()
	
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
	
	m.featuresList = list.New(items, delegate, m.width/4, listHeight)
	m.featuresList.Title = "Features"
	m.featuresList.SetShowStatusBar(false)
	m.featuresList.SetFilteringEnabled(false)
	m.featuresList.SetShowHelp(false)
	
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

// initEndpointsList initializes the endpoints list
func (m *Model) initEndpointsList() {
	delegate := list.NewDefaultDelegate()
	
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
				
				// Sort responses alphabetically
				for i := 0; i < len(allResponses); i++ {
					for j := i + 1; j < len(allResponses); j++ {
						if allResponses[i] > allResponses[j] {
							allResponses[i], allResponses[j] = allResponses[j], allResponses[i]
						}
					}
				}
				
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
	
	// Create the list with proper dimensions
	listHeight := m.height - 6 // Account for header and footer
	if listHeight < 1 {
		listHeight = 10 // Default if height not set yet
	}
	
	m.endpointsList = list.New(items, delegate, 3*m.width/4, listHeight)
	m.endpointsList.Title = fmt.Sprintf("Endpoints (%s)", m.selectedFeature)
	m.endpointsList.SetShowStatusBar(false)
	m.endpointsList.SetFilteringEnabled(false)
	m.endpointsList.SetShowHelp(false)
}

// updateEndpointsList updates the endpoints list based on the selected feature
func (m *Model) updateEndpointsList() {
	// Save current selection index
	currentIndex := m.endpointsList.Index()
	
	// Create new items without recreating the entire list
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
				
				// Sort responses alphabetically
				for i := 0; i < len(allResponses); i++ {
					for j := i + 1; j < len(allResponses); j++ {
						if allResponses[i] > allResponses[j] {
							allResponses[i], allResponses[j] = allResponses[j], allResponses[i]
						}
					}
				}
				
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
	
	// Update just the items, not the entire list
	m.endpointsList.SetItems(items)
	m.endpointsList.Title = fmt.Sprintf("Endpoints (%s)", m.selectedFeature)
	
	// Restore selection if possible
	if currentIndex < len(m.endpointsList.Items()) {
		m.endpointsList.Select(currentIndex)
	}
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
							
							// Sort responses alphabetically
							for i := 0; i < len(allResponses); i++ {
								for j := i + 1; j < len(allResponses); j++ {
									if allResponses[i] > allResponses[j] {
										allResponses[i], allResponses[j] = allResponses[j], allResponses[i]
									}
								}
							}
							
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
		
		m.featuresList.SetSize(m.width/4, listHeight)
		m.endpointsList.SetSize(3*m.width/4, listHeight)
		
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
		case key.Matches(msg, m.keyMap.Tab):
			m.togglePanel()
		case key.Matches(msg, m.keyMap.Help):
			m.activeDialog = HelpDialog
			// Initialize dialog content
			m.dialogTitle = "Mockoho Help"
			m.dialogContent = ""
			return m, nil
		case key.Matches(msg, m.keyMap.New):
			if m.activePanel == FeaturesPanel {
				m.showNewFeatureDialog()
			} else if len(m.selectedFeature) > 0 {
				m.showNewEndpointDialog()
			} else {
				// Can't create endpoint without a selected feature
				return m, nil
			}
			return m, nil
		case key.Matches(msg, m.keyMap.Delete):
			m.showDeleteConfirmDialog()
			return m, nil
		case key.Matches(msg, m.keyMap.Proxy):
			m.showProxyConfigDialog()
			return m, nil
		case key.Matches(msg, m.keyMap.Server):
			return m, m.toggleServer
		case key.Matches(msg, m.keyMap.Reload):
			return m, m.reloadConfig
		case key.Matches(msg, m.keyMap.Toggle):
			if m.activePanel == EndpointsPanel {
				return m, m.toggleEndpoint()
			}
		case key.Matches(msg, m.keyMap.Response):
			if m.activePanel == EndpointsPanel {
				return m, m.cycleResponse()
			}
		case key.Matches(msg, m.keyMap.Open):
			return m, m.openInEditor
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

// togglePanel toggles between the features and endpoints panels
func (m *Model) togglePanel() {
	if m.activePanel == FeaturesPanel {
		m.activePanel = EndpointsPanel
	} else {
		m.activePanel = FeaturesPanel
	}
}

// toggleServer toggles the server on/off
func (m *Model) toggleServer() tea.Msg {
	if m.Server.IsRunning() {
		if err := m.Server.Stop(); err != nil {
			return fmt.Errorf("failed to stop server: %v", err)
		}
		return nil
	} else {
		if err := m.Server.Start(); err != nil {
			return fmt.Errorf("failed to start server: %v", err)
		}
		return nil
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
		if m.activePanel != EndpointsPanel {
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
		if m.activePanel != EndpointsPanel {
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
	
	if m.activePanel == FeaturesPanel {
		item, ok := m.featuresList.SelectedItem().(featureItem)
		if !ok {
			return fmt.Errorf("no feature selected")
		}
		
		filePath = fmt.Sprintf("%s/%s.json", m.Config.BaseDir, item.name)
		line = 1
	} else {
		endpoint, ok := m.endpointsList.SelectedItem().(endpointItem)
		if !ok {
			return fmt.Errorf("no endpoint selected")
		}
		
		filePath = fmt.Sprintf("%s/%s.json", m.Config.BaseDir, m.selectedFeature)
		
		// Find the line number of the endpoint
		// This is a simple approximation
		line = 10 // Default line number
		
		// We could use endpoint.id to find the actual line number in the file
		_ = endpoint.id // Use the variable to avoid unused variable error
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
	
	args := make([]string, len(m.Config.Global.Editor.Args))
	copy(args, m.Config.Global.Editor.Args)
	
	// Replace placeholders in args
	for i, arg := range args {
		args[i] = strings.ReplaceAll(arg, "{file}", filePath)
		args[i] = strings.ReplaceAll(arg, "{line}", fmt.Sprintf("%d", line))
	}
	
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
				fmt.Println("No input focused, focusing first input")
				return m, nil
			}
			
			// Blur the current input
			m.textInputs[focusedIndex].Blur()
			
			// Focus the next input (or wrap around to the first)
			nextIndex := (focusedIndex + 1) % len(m.textInputs)
			m.textInputs[nextIndex].Focus()
			
			fmt.Printf("Tab pressed: Moving focus from input %d to input %d\n",
				focusedIndex, nextIndex)
			
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
				fmt.Println("No input focused, focusing last input")
				return m, nil
			}
			
			// Blur the current input
			m.textInputs[focusedIndex].Blur()
			
			// Focus the previous input (or wrap around to the last)
			prevIndex := (focusedIndex - 1 + len(m.textInputs)) % len(m.textInputs)
			m.textInputs[prevIndex].Focus()
			
			fmt.Printf("Shift+Tab pressed: Moving focus from input %d to input %d\n",
				focusedIndex, prevIndex)
			
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
				fmt.Println("No input focused after update, focusing first input")
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