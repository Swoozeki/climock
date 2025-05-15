# Climock Developer Guide

This guide provides technical information for developers who want to understand or contribute to the Climock codebase.

## Table of Contents

1. [Project Structure](#project-structure)
2. [TUI Architecture](#tui-architecture)
3. [Core Components](#core-components)
4. [Integration with Core Services](#integration-with-core-services)
5. [Best Practices](#best-practices)
6. [Common Patterns](#common-patterns)

## Project Structure

Climock is organized into the following main directories:

```
climock/
  ├── cmd/                  # Command-line entry points
  │   └── climock/          # Main application
  ├── docs/                 # Documentation
  ├── examples/             # Example code and configurations
  ├── internal/             # Internal packages
  │   ├── config/           # Configuration management
  │   ├── logger/           # Logging functionality
  │   ├── mock/             # Mock server implementation
  │   ├── proxy/            # Proxy server implementation
  │   ├── server/           # HTTP server implementation
  │   └── ui/               # Terminal user interface
  └── mocks/                # Default mock configurations
```

## TUI Architecture

Climock's TUI is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, a Go library for building terminal applications using The Elm Architecture (TEA). The UI is styled using [Lip Gloss](https://github.com/charmbracelet/lipgloss), which provides tools for styling terminal applications.

### Core Components

#### Model (`internal/ui/model.go`)

The `Model` struct is the central data structure that holds the application state:

```go
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
}
```

Key components:

- References to core services (Config, MockManager, ProxyManager, Server)
- UI state (active panel, list models, dimensions)
- Dialog state (active dialog, inputs, callbacks)

#### View (`internal/ui/view.go`)

The `View` function renders the UI based on the current model state:

```go
func (m *Model) View() string {
    // If a dialog is active, render it
    if m.activeDialog != NoDialog {
        return m.renderDialog()
    }

    // Render the main UI
    var sb strings.Builder

    // Header
    sb.WriteString(m.renderHeader())
    sb.WriteString("\n")

    // Panel titles
    sb.WriteString(m.renderPanelTitles())
    sb.WriteString("\n")

    // Lists
    sb.WriteString(m.renderLists())
    sb.WriteString("\n")

    // Footer
    sb.WriteString(m.renderFooter())

    return sb.String()
}
```

The view is composed of several components:

- Header (showing server status and proxy target)
- Panel titles
- Lists (features and endpoints)
- Footer (showing keyboard shortcuts)
- Dialogs (when active)

#### Update (`internal/ui/model.go`)

The `Update` function handles state changes based on messages:

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        // Handle window size changes
        // ...
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
        // ... other key handlers
        }
    }

    // Update the active list
    if m.activePanel == FeaturesPanel {
        var listCmd tea.Cmd
        m.featuresList, listCmd = m.featuresList.Update(msg)
        cmds = append(cmds, listCmd)

        // Update selected feature when list selection changes
        // ...
    } else {
        var listCmd tea.Cmd
        m.endpointsList, listCmd = m.endpointsList.Update(msg)
        cmds = append(cmds, listCmd)
    }

    return m, tea.Batch(cmds...)
}
```

The update function handles:

- Window size changes
- Keyboard input
- List updates
- Dialog interactions

### UI Panels

#### Features Panel

The Features Panel displays a list of available features (groups of endpoints):

- Located on the left side of the screen
- Shows feature names
- Allows selection of a feature to view its endpoints
- Supports creating new features and deleting existing ones

#### Endpoints Panel

The Endpoints Panel displays a list of endpoints for the selected feature:

- Located on the right side of the screen
- Shows endpoint method, path, and active status (🟢/🔴)
- Shows available responses with the default response marked (★)
- Supports creating new endpoints, deleting endpoints, toggling active state, and cycling through responses

### Dialog System

Climock uses a modal dialog system for user input and confirmation:

#### Dialog Types

```go
const (
    NoDialog DialogType = iota
    HelpDialog
    NewFeatureDialog
    NewEndpointDialog
    DeleteConfirmDialog
    ProxyConfigDialog
)
```

#### Dialog Components

- **Help Dialog**: Shows keyboard shortcuts and usage instructions
- **New Feature Dialog**: Input form for creating a new feature
- **New Endpoint Dialog**: Input form for creating a new endpoint
- **Delete Confirm Dialog**: Confirmation dialog for deleting features or endpoints
- **Proxy Config Dialog**: Input form for configuring the proxy target

### Keyboard Navigation

Climock uses a keyboard-driven interface with the following key bindings:

```go
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
        Tab: key.NewBinding(
            key.WithKeys("tab"),
            key.WithHelp("tab", "switch panel"),
        ),
        // ... other key bindings
    }
}
```

### List Components

Climock uses Bubble Tea's list component for both the Features and Endpoints panels:

#### Feature Item

```go
type featureItem struct {
    name string
}
```

#### Endpoint Item

```go
type endpointItem struct {
    id              string
    method          string
    path            string
    active          bool
    defaultResponse string
    responses       []string
}
```

### Styling

Climock uses Lip Gloss for styling terminal UI components:

```go
func (m *Model) renderHeader() string {
    headerStyle := lipgloss.NewStyle().
        BorderStyle(lipgloss.NormalBorder()).
        BorderBottom(true).
        Width(m.width)

    serverStatus := "Stopped"
    if m.Server.IsRunning() {
        serverStatus = fmt.Sprintf("Running (%s)", m.Server.GetAddress())
    }

    proxyTarget := m.ProxyManager.GetTargetURL()
    header := fmt.Sprintf("Server: %s | Proxy: %s", serverStatus, proxyTarget)

    return headerStyle.Render("Climock - " + header)
}
```

### Initialization and Lifecycle

#### Initialization

```go
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
    }

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
```

#### Init Function

```go
func (m *Model) Init() tea.Cmd {
    // Return commands to enter alt screen and get window size
    return tea.Batch(
        tea.EnterAltScreen,
        func() tea.Msg {
            return tea.WindowSizeMsg{
                Width:  m.width,
                Height: m.height,
            }
        },
    )
}
```

## Integration with Core Services

The TUI integrates with the following core services:

### Config Service

- Loads and saves configuration files
- Provides access to feature and endpoint configurations

### Mock Manager

- Creates, updates, and deletes features and endpoints
- Toggles endpoint active state
- Sets default responses

### Proxy Manager

- Updates proxy target
- Provides proxy target URL for display

### Server

- Starts and stops the mock server
- Provides server status and address for display
- Reloads configuration when changes are made

## Best Practices

When extending or modifying the TUI, follow these best practices:

1. **State Management**: Keep all state in the Model struct and update it through the Update function
2. **View Separation**: Keep view logic separate from state management
3. **Responsive Design**: Handle window size changes to ensure the UI looks good at different terminal sizes
4. **Error Handling**: Display errors to the user in a clear and non-disruptive way
5. **Keyboard Navigation**: Ensure all functionality is accessible via keyboard shortcuts
6. **Visual Feedback**: Provide clear visual feedback for user actions

## Common Patterns

### Adding a New Dialog

1. Add a new dialog type to the DialogType enum
2. Create a show function (e.g., `showNewDialog()`)
3. Add a render function if needed (or use an existing one)
4. Add keyboard handling in the updateDialog function
5. Add a command function to handle the dialog's action

### Adding a New Keyboard Shortcut

1. Add a new binding to the KeyMap struct
2. Add the binding to DefaultKeyMap()
3. Add the binding to ShortHelp() and/or FullHelp()
4. Handle the key press in the Update function
