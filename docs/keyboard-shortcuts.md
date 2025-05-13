# Mockoho Keyboard Shortcuts Reference

This document provides a comprehensive reference for all keyboard shortcuts available in the Mockoho CLI interface.

## Navigation Shortcuts

| Key     | Action       | Description                                  |
| ------- | ------------ | -------------------------------------------- |
| `Tab`   | Switch Panel | Switch between Features and Endpoints panels |
| `↑`     | Move Up      | Move up in the current panel                 |
| `↓`     | Move Down    | Move down in the current panel               |
| `Enter` | Select       | Select a feature or endpoint                 |

## Feature & Endpoint Management

| Key | Action   | Description                                                             |
| --- | -------- | ----------------------------------------------------------------------- |
| `n` | New      | Create new feature (in Features panel) or endpoint (in Endpoints panel) |
| `d` | Delete   | Delete selected feature or endpoint                                     |
| `t` | Toggle   | Toggle endpoint active/inactive                                         |
| `r` | Response | Cycle through available responses for the endpoint                      |

## Server & Configuration

| Key      | Action | Description                               |
| -------- | ------ | ----------------------------------------- |
| `s`      | Server | Start/stop the mock server                |
| `p`      | Proxy  | Configure proxy target                    |
| `o`      | Open   | Open configuration file in default editor |
| `Ctrl+r` | Reload | Reload configurations from disk           |

## Search & Help

| Key        | Action | Description                         |
| ---------- | ------ | ----------------------------------- |
| `/`        | Search | Search for endpoints                |
| `h` or `?` | Help   | Show help screen with all shortcuts |

## Dialog Navigation

| Key     | Action  | Description           |
| ------- | ------- | --------------------- |
| `Enter` | Confirm | Confirm dialog action |
| `Esc`   | Cancel  | Cancel dialog         |

## Application Control

| Key             | Action | Description          |
| --------------- | ------ | -------------------- |
| `q` or `Ctrl+c` | Quit   | Exit the application |

## Tips for Efficient Usage

### Quick Feature Creation

1. Press `Tab` to ensure you're in the Features panel
2. Press `n` to create a new feature
3. Enter the feature name and press `Enter`

### Quick Endpoint Creation

1. Select a feature in the Features panel
2. Press `Tab` to switch to the Endpoints panel
3. Press `n` to create a new endpoint
4. Fill in the endpoint details and press `Enter`

### Rapid Response Switching

1. Select an endpoint in the Endpoints panel
2. Press `r` repeatedly to cycle through available responses
3. The current response is indicated with a star (★) in the endpoint description

### Efficient Configuration Editing

1. Select a feature or endpoint
2. Press `o` to open the configuration file in your editor
3. Make changes and save
4. Press `Ctrl+r` to reload the configuration

### Server Management

1. Press `s` to start the server
2. Test your endpoints
3. Press `s` again to stop the server when done

### Proxy Configuration

1. Press `p` to open the proxy configuration dialog
2. Enter the target URL for proxying requests
3. Press `Enter` to confirm

### Search Functionality

1. Press `/` to activate search
2. Type your search query
3. Press `Enter` to search
4. Use `↑/↓` to navigate through search results

## Customizing Keyboard Shortcuts

Currently, keyboard shortcuts are hardcoded in the application. Future versions may support customization through a configuration file.

## Dialog Keyboard Navigation

When in a dialog:

- Use `Tab` to navigate between input fields
- Use `Enter` to confirm
- Use `Esc` to cancel

## Common Workflows

### Creating a New Mock API

1. Press `n` in the Features panel to create a new feature
2. Press `Tab` to switch to the Endpoints panel
3. Press `n` to create a new endpoint
4. Press `o` to open the configuration file and customize responses
5. Press `s` to start the server and test your mock API

### Toggling Between Mock and Proxy

1. Select an endpoint in the Endpoints panel
2. Press `t` to toggle the endpoint active/inactive
3. When inactive, requests will be proxied to the real backend

### Managing Multiple Response Scenarios

1. Select an endpoint in the Endpoints panel
2. Press `o` to open the configuration file
3. Add multiple responses in the JSON configuration
4. Save and close the file
5. Press `Ctrl+r` to reload the configuration
6. Press `r` to cycle through the available responses
