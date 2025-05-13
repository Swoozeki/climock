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

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Tab, k.Enter},
		{k.Toggle, k.Response, k.Open, k.New, k.Delete},
		{k.Proxy, k.Server, k.Quit, k.Help, k.Search, k.Reload},
	}
}