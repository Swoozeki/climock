# Mockoho: Mock Server System

Mockoho is a mock server for development. It'll act as proxy to the real server where mock endpoints are not defined. The primary interface is the interactive CLI, where you can add new features and mock endpoints, and easily toggle any of them for quick and easy mocking.

## Features

- **Mock API Endpoints**: Create and manage mock API endpoints with custom responses
- **Multiple Response Support**: Configure multiple responses per endpoint and easily switch between them
- **Proxy Mode**: Automatically proxy requests to a real backend for endpoints without mocks
- **Ultra-Compact CLI Tool**: Intuitive keyboard-driven interface for managing mock configurations
- **External Editor Integration**: Open configuration files in your preferred editor

## Installation

### Option 1: Using Homebrew (Recommended)

```bash
# Add our private tap (you'll be prompted for your GitHub credentials)
brew tap kohofinancial/tap https://github.com/kohofinancial/homebrew-tap.git

# Install mockoho
brew install mockoho
```

### Option 2: Building from Source

```bash
# Clone the repository
git clone https://github.com/kohofinancial/mockoho.git
cd mockoho

# Build the application
go build -o mockoho ./cmd/mockoho
```

## Usage

After installation, you can start Mockoho with:

```bash
# Run with default configuration
mockoho

# Specify a custom configuration directory
mockoho --config /path/to/your/mocks

# Run in server-only mode (without TUI)
mockoho server --config /path/to/your/mocks
```

The Terminal User Interface (TUI) will launch, allowing you to manage mock configurations using keyboard shortcuts. Your mock API will be available at `http://localhost:3000/api/...`

## Quick Start

1. Run the Mockoho CLI tool:

```bash
mockoho
```

2. Use the keyboard shortcuts to navigate and manage mock configurations:

   - `←/→` to switch between Features and Endpoints panels
   - `↑/↓` to navigate up/down in the current panel
   - `n` to add new feature or endpoint
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

## License

MIT
