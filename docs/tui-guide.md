# Mockoho TUI User Guide

This guide explains how to use the Terminal User Interface (TUI) of Mockoho, a mock server system for API development and testing.

## Getting Started

To launch the Mockoho TUI, run:

```bash
./mockoho
```

This will open the interactive terminal interface.

## Understanding the Interface

The Mockoho interface consists of two main panels:

1. **Features Panel** (left): Lists all available features (groups of endpoints)
2. **Endpoints Panel** (right): Lists all endpoints for the selected feature

At the top, you'll see a header showing the server status and proxy target. At the bottom, you'll see a footer showing available keyboard shortcuts.

```
┌─Mockoho - Server: Stopped | Proxy: https://api.real-server.com─────────────────────────────────────┐
│                                                                                                     │
├─Features───────────────────┬─Endpoints (users)──────────────────────────────────────────────────────┤
│                            │                                                                        │
│ > users                    │ > GET /api/users/:id 🟢                                                │
│   products                 │   [★standard | premium | error]                                        │
│   auth                     │                                                                        │
│                            │ > POST /api/users 🟢                                                   │
│                            │   [★success | validation-error]                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
├────────────────────────────┴────────────────────────────────────────────────────────────────────────┤
│ t toggle  r response  o open  n new  d delete  p proxy  s server  q quit  h help                    │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Basic Navigation

- **Tab**: Switch between Features and Endpoints panels
- **↑/↓**: Navigate up/down in the current panel
- **Enter**: Select a feature or endpoint
- **q**: Quit the application

## Working with Features

### Viewing Features

The Features panel shows all available features. Each feature represents a group of related endpoints.

### Creating a New Feature

1. Press **Tab** to ensure you're in the Features panel
2. Press **n** to open the new feature dialog
3. Enter a name for your feature (e.g., "products")
4. Press **Enter** to confirm

### Deleting a Feature

1. Select the feature you want to delete in the Features panel
2. Press **d** to open the delete confirmation dialog
3. Press **Enter** to confirm deletion

## Working with Endpoints

### Viewing Endpoints

The Endpoints panel shows all endpoints for the selected feature. Each endpoint shows:

- HTTP method (GET, POST, PUT, etc.)
- Path (e.g., /api/users/:id)
- Active status (🟢 for active, 🔴 for inactive)
- Available responses with the default response marked with a star (★)

### Creating a New Endpoint

1. Select a feature in the Features panel
2. Press **Tab** to switch to the Endpoints panel
3. Press **n** to open the new endpoint dialog
4. Fill in the required fields:
   - **Endpoint ID**: A unique identifier (e.g., "get-products")
   - **Method**: HTTP method (e.g., "GET")
   - **Path**: API path (e.g., "/api/products")
5. Press **Enter** to confirm

### Toggling Endpoint Active State

1. Select an endpoint in the Endpoints panel
2. Press **t** to toggle between active (🟢) and inactive (🔴)

When an endpoint is inactive, requests to that endpoint will be proxied to the real backend (if configured).

### Cycling Through Responses

1. Select an endpoint in the Endpoints panel
2. Press **r** to cycle through the available responses
3. The current default response is marked with a star (★)

### Editing Endpoint Configuration

1. Select an endpoint in the Endpoints panel
2. Press **o** to open the configuration file in your default editor
3. Make changes to the JSON configuration
4. Save and close the editor
5. Press **Ctrl+r** to reload the configuration

## Server Management

### Starting the Server

Press **s** to start the mock server. The header will update to show "Server: Running (localhost:3000)".

### Stopping the Server

Press **s** again to stop the server. The header will update to show "Server: Stopped".

## Proxy Configuration

### Setting the Proxy Target

1. Press **p** to open the proxy configuration dialog
2. Enter the target URL (e.g., "https://api.real-server.com")
3. Press **Enter** to confirm

## Help and Information

### Viewing Help

Press **h** to open the help dialog, which shows all available keyboard shortcuts:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                     │
│                                          Mockoho Help                                               │
│                                                                                                     │
│ Navigation:                                                                                         │
│   Tab       - Switch between Features and Endpoints panels                                          │
│   ↑/↓       - Navigate up/down in the current panel                                                 │
│   Enter     - Select a feature or endpoint                                                          │
│                                                                                                     │
│ Actions:                                                                                            │
│   t         - Toggle endpoint active/inactive                                                       │
│   r         - Cycle through available responses (sets as default)                                   │
│   o         - Open configuration file in default editor                                             │
│   n         - Create new endpoint or feature                                                        │
│   d         - Delete selected endpoint or feature                                                   │
│   p         - Change proxy target                                                                   │
│   s         - Start/stop server                                                                     │
│   q         - Quit application                                                                      │
│   h         - Show this help screen                                                                 │
│   /         - Search for endpoints                                                                  │
│   Ctrl+r    - Reload configurations from disk                                                       │
│                                                                                                     │
│ Press Esc or any key to return...                                                                   │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Searching

1. Press **/** to activate search mode
2. Type your search query
3. Press **Enter** to search
4. Use **↑/↓** to navigate through search results

## Common Workflows

### Creating a New Mock API

1. Press **n** in the Features panel to create a new feature
2. Press **Tab** to switch to the Endpoints panel
3. Press **n** to create a new endpoint
4. Press **o** to open the configuration file and customize responses
5. Press **s** to start the server and test your mock API

### Toggling Between Mock and Proxy

1. Select an endpoint in the Endpoints panel
2. Press **t** to toggle the endpoint active/inactive
3. When inactive, requests will be proxied to the real backend

### Managing Multiple Response Scenarios

1. Select an endpoint in the Endpoints panel
2. Press **o** to open the configuration file
3. Add multiple responses in the JSON configuration
4. Save and close the file
5. Press **Ctrl+r** to reload the configuration
6. Press **r** to cycle through the available responses

## Dialog Navigation

When in a dialog:

- Use **Tab** to navigate between input fields
- Use **Enter** to confirm
- Use **Esc** to cancel

## Tips and Tricks

### Efficient Configuration Editing

1. Select a feature or endpoint
2. Press **o** to open the configuration file in your editor
3. Make changes and save
4. Press **Ctrl+r** to reload the configuration

### Rapid Response Switching

1. Select an endpoint in the Endpoints panel
2. Press **r** repeatedly to cycle through available responses
3. The current response is indicated with a star (★) in the endpoint description

### Server Management

1. Press **s** to start the server
2. Test your endpoints
3. Press **s** again to stop the server when done

## Troubleshooting

### Changes Not Reflected

- Press **Ctrl+r** to reload the configuration
- Check for JSON syntax errors in your configuration files

### Server Won't Start

- Check if the port is already in use
- Verify that the configuration files are valid JSON

### Proxy Not Working

- Verify that the proxy target is correct and accessible
- Check that the endpoint is inactive or not defined
