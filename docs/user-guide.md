# Mockoho User Guide

A powerful mock server system for frontend development without a ready backend and for easily updating responses for demo purposes.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Usage](#basic-usage)
3. [Terminal UI & Keyboard Shortcuts](#terminal-user-interface-and-shortcuts)
4. [Configuration Reference](#configuration-reference)
5. [Advanced Usage](#advanced-usage)
6. [Troubleshooting](#troubleshooting)

## Getting Started

### Installation

**Option 1: Download Binary (Recommended)**

1. Download the latest release from [GitHub Releases](https://github.com/kohofinancial/mockoho/releases)
2. Extract and move the binary to a location in your PATH

**Option 2: Using Go Install**

```bash
go install github.com/kohofinancial/mockoho@latest
```

### Running the Application

```bash
# Basic usage
mockoho

# With custom config directory
mockoho --config /path/to/your/mocks

# Server-only mode (no UI)
mockoho server --config /path/to/your/mocks
```

### Interface Overview

Mockoho has a keyboard-driven interface with two main panels:

- **Features Panel** (left): Lists all available features (groups of endpoints)
- **Endpoints Panel** (right): Lists all endpoints for the selected feature

```
┌─Mockoho - Server: Stopped | Proxy: https://api.real-server.com─────────────────┐
│                                                                                 │
├─Features───────────────────┬─Endpoints (users)──────────────────────────────────┤
│                            │                                                    │
│ > users                    │ > GET /api/users/:id 🟢                            │
│   products                 │   [★standard | premium | error]                    │
│   auth                     │                                                    │
│                            │ > POST /api/users 🟢                               │
│                            │   [★success | validation-error]                    │
│                            │                                                    │
├────────────────────────────┴────────────────────────────────────────────────────┤
│ t toggle  r response  o open  n new  d delete  p proxy  s server  q quit  h help│
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Basic Usage

### Creating Mocks

1. **Create a Feature** (Tab to Features panel → n → enter name)
2. **Create an Endpoint** (Tab to Endpoints panel → n → fill in fields)
3. **Edit Configuration** (Select endpoint → o → modify JSON → save)
4. **Start Server** (s)
5. **Test Your Mock** (`curl http://localhost:3000/your-endpoint`)

### Working with Responses

- **Cycle Responses**: Select endpoint → r
- **Toggle Active/Inactive**: Select endpoint → t
- **Configure Proxy**: p → enter target URL

## Terminal User Interface and Shortcuts

### Navigation

- **Tab**: Switch between panels
- **↑/↓**: Navigate up/down
- **Enter**: Select item

### Key Actions

| Key    | Action   | Description                     |
| ------ | -------- | ------------------------------- |
| n      | New      | Create feature/endpoint         |
| d      | Delete   | Delete feature/endpoint         |
| t      | Toggle   | Toggle endpoint active/inactive |
| r      | Response | Cycle through responses         |
| s      | Server   | Start/stop server               |
| p      | Proxy    | Configure proxy target          |
| o      | Open     | Open config in editor           |
| Ctrl+r | Reload   | Reload configurations           |
| /      | Search   | Search for endpoints            |
| h or ? | Help     | Show help screen                |
| q      | Quit     | Exit application                |

## Configuration Reference

### Global Configuration (`mocks/config.json`)

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
  }
}
```

### Endpoint Configuration

```json
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
      "delay": 0
    }
  }
}
```

### Template Variables

| Variable          | Description                  | Example                                                                          |
| ----------------- | ---------------------------- | -------------------------------------------------------------------------------- |
| `{{params.name}}` | Path parameter value         | If path is `/api/users/:id`, then `{{params.id}}` is replaced with the actual ID |
| `{{now}}`         | Current timestamp (ISO 8601) | `"2023-05-13T14:30:00.000Z"`                                                     |

### File Structure

```
mocks/
  ├── config.json     # Global configuration
  ├── users.json      # Feature configuration for users
  ├── products.json   # Feature configuration for products
  └── ...             # Other feature configurations
```

## Advanced Usage

### Response Delay

Add a delay to simulate network latency:

```json
"delay": 2000  // 2 seconds delay
```

### Custom Headers

```json
"headers": {
  "Content-Type": "application/json",
  "X-Custom-Header": "Custom Value",
  "X-Rate-Limit": "100"
}
```

### Path Parameters

```json
{
  "path": "/api/users/:userId/files/:fileId",
  "responses": {
    "standard": {
      "body": {
        "userId": "{{params.userId}}",
        "fileId": "{{params.fileId}}"
      }
    }
  }
}
```

## Troubleshooting

| Problem               | Solution                                                                  |
| --------------------- | ------------------------------------------------------------------------- |
| Server won't start    | Check if port is in use; verify JSON is valid                             |
| Changes not reflected | Press Ctrl+r to reload; check for JSON syntax errors                      |
| Proxy not working     | Verify proxy target is correct and accessible; check endpoint is inactive |
