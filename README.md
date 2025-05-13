# Mockoho: Mock Server System

Mockoho is a powerful mock server system designed to facilitate frontend development without a ready backend and to easily update responses for demo purposes.

## Features

- **Mock API Endpoints**: Create and manage mock API endpoints with custom responses
- **Multiple Response Support**: Configure multiple responses per endpoint and easily switch between them
- **Proxy Mode**: Automatically proxy requests to a real backend for endpoints without mocks
- **Ultra-Compact CLI Tool**: Intuitive keyboard-driven interface for managing mock configurations
- **External Editor Integration**: Open configuration files in your preferred editor

## Installation

### Prerequisites

- Go 1.18 or higher

### Building from Source

```bash
# Clone the repository
git clone https://github.com/mockoho/mockoho.git
cd mockoho

# Build the application
go build -o mockoho ./cmd/mockoho
```

## Quick Start

1. Run the Mockoho CLI tool:

```bash
./mockoho
```

2. Use the keyboard shortcuts to navigate and manage mock configurations:

   - `Tab` to switch between Features and Endpoints panels
   - `↑/↓` to navigate up/down in the current panel
   - `t` to toggle endpoint active/inactive
   - `r` to cycle through available responses
   - `s` to start/stop the server
   - `h` to show help screen with all shortcuts

3. Access your mock API at `http://localhost:3000/api/...`

## Configuration

Mockoho uses JSON files for configuration, organized by feature name:

```
mocks/
  ├── config.json     # Global configuration
  ├── users.json      # User-related endpoints
  ├── products.json   # Product-related endpoints
  └── ...
```

### Global Configuration (config.json)

```json
{
  "proxyConfig": {
    "target": "https://api.real-server.com",
    "changeOrigin": true,
    "pathRewrite": {
      "^/api": ""
    }
  },
  "serverConfig": {
    "port": 3000,
    "host": "localhost"
  },
  "editor": {
    "command": "code",
    "args": ["-g", "{file}:{line}"]
  }
}
```

### Feature-Based Mock Definition (e.g., users.json)

```json
{
  "feature": "users",
  "endpoints": [
    {
      "id": "get-user-profile",
      "method": "GET",
      "path": "/api/users/:id",
      "active": true,
      "defaultResponse": "standard",
      "responses": {
        "standard": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "{{params.id}}",
            "name": "John Doe",
            "email": "john@example.com"
          },
          "delay": 200
        },
        "premium": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "{{params.id}}",
            "name": "John Doe",
            "email": "john@example.com",
            "premium": true,
            "memberSince": "2020-01-01"
          },
          "delay": 200
        },
        "error": {
          "status": 404,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "User not found",
            "code": "USER_NOT_FOUND"
          },
          "delay": 100
        }
      }
    }
  ]
}
```

## Template Variables

Mockoho supports template variables in response bodies:

- `{{params.id}}` - Path parameter value (e.g., `:id` in `/api/users/:id`)
- `{{now}}` - Current timestamp in ISO 8601 format

## Command Line Options

```
Usage:
  mockoho [flags]
  mockoho [command]

Available Commands:
  help        Help about any command
  server      Start the mock server without the UI

Flags:
  -c, --config string   Directory containing mock configurations (default "mocks")
  -h, --help            help for mockoho
```

## Keyboard Shortcuts

| Key    | Action        | Description                                        |
| ------ | ------------- | -------------------------------------------------- |
| Tab    | Switch Panel  | Switch between Features and Endpoints panels       |
| ↑/↓    | Navigate List | Move up/down in the current panel                  |
| Enter  | Select Item   | Select a feature or endpoint                       |
| t      | Toggle        | Toggle endpoint active/inactive                    |
| r      | Response      | Cycle through available responses for the endpoint |
| o      | Open          | Open configuration file in default editor          |
| n      | New           | Create new endpoint or feature                     |
| d      | Delete        | Delete selected endpoint or feature                |
| p      | Proxy         | Change proxy target                                |
| s      | Server        | Start/stop server                                  |
| q      | Quit          | Exit the application                               |
| h      | Help          | Show help screen with all shortcuts                |
| /      | Search        | Search for endpoints                               |
| Ctrl+r | Reload        | Reload configurations from disk                    |

## License

MIT
