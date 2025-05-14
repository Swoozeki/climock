package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

// Panel represents a UI panel
type Panel int

const (
	FeaturesPanel Panel = iota
	EndpointsPanel
)

// DialogType represents the type of dialog
type DialogType int

const (
	NoDialog DialogType = iota
	HelpDialog
	NewFeatureDialog
	NewEndpointDialog
	DeleteConfirmDialog
	ProxyConfigDialog
)

// KeyMap defines the keybindings for the UI
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Tab          key.Binding
	Enter        key.Binding
	Toggle       key.Binding
	Response     key.Binding
	Open         key.Binding
	New          key.Binding
	Delete       key.Binding
	Proxy        key.Binding
	Server       key.Binding
	Quit         key.Binding
	Help         key.Binding
	Search       key.Binding
	Reload       key.Binding
	Escape       key.Binding
	Confirm      key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "move right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle endpoint"),
		),
		Response: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "cycle response"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in editor"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new item"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete item"),
		),
		Proxy: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "proxy config"),
		),
		Server: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start/stop server"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("h", "?"),
			key.WithHelp("h", "help"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Reload: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reload"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Toggle, k.Response, k.Open, k.New,
		k.Delete, k.Proxy, k.Server, k.Quit, k.Help,
	}
}

// PanelKeyMap is a wrapper around KeyMap that provides panel-specific keybindings
type PanelKeyMap struct {
	keyMap      KeyMap
	activePanel Panel
}

// NewPanelKeyMap creates a new PanelKeyMap
func NewPanelKeyMap(keyMap KeyMap, activePanel Panel) PanelKeyMap {
	return PanelKeyMap{
		keyMap:      keyMap,
		activePanel: activePanel,
	}
}

// ShortHelp returns panel-specific keybindings for the mini help view
func (pk PanelKeyMap) ShortHelp() []key.Binding {
	// Common shortcuts for both panels
	commonBindings := []key.Binding{
		pk.keyMap.Open, pk.keyMap.New, pk.keyMap.Delete,
		pk.keyMap.Server, pk.keyMap.Proxy, pk.keyMap.Quit, pk.keyMap.Help,
	}
	
	// Panel-specific shortcuts
	if pk.activePanel == FeaturesPanel {
		return commonBindings
	} else { // EndpointsPanel
		return append([]key.Binding{pk.keyMap.Toggle, pk.keyMap.Response}, commonBindings...)
	}
}

// ShortHelpInRows returns panel-specific keybindings split into two rows
// for consistent footer height
func (pk PanelKeyMap) ShortHelpInRows() [][]key.Binding {
	// Item-specific shortcuts on top row
	row1 := []key.Binding{}
	
	// Panel-specific shortcuts
	if pk.activePanel == EndpointsPanel {
		// Endpoint-specific actions
		row1 = append(row1, pk.keyMap.Toggle, pk.keyMap.Response)
	}
	
	// Common item actions
	row1 = append(row1, pk.keyMap.Open, pk.keyMap.New, pk.keyMap.Delete)
	
	// General application shortcuts on bottom row
	row2 := []key.Binding{
		pk.keyMap.Server, pk.keyMap.Proxy, pk.keyMap.Quit, pk.keyMap.Help,
	}
	
	return [][]key.Binding{row1, row2}
}

// FullHelp returns keybindings for the expanded help view
func (pk PanelKeyMap) FullHelp() [][]key.Binding {
	return pk.keyMap.FullHelp()
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Tab, k.Enter},
		{k.Toggle, k.Response, k.Open, k.New, k.Delete},
		{k.Proxy, k.Server, k.Quit, k.Help, k.Search, k.Reload},
	}
}