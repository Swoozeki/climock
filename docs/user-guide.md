# Mockoho User Guide

This comprehensive guide covers everything you need to know about using Mockoho, a powerful mock server system designed to facilitate frontend development without a ready backend and to easily update responses for demo purposes.

## Table of Contents

1. [Getting Started](#getting-started)

   - [Installation](#installation)
   - [Running the Application](#running-the-application)
   - [Understanding the Interface](#understanding-the-interface)

2. [Basic Usage](#basic-usage)

   - [Creating Your First Mock](#creating-your-first-mock)
   - [Working with Multiple Responses](#working-with-multiple-responses)
   - [Configuring the Proxy](#configuring-the-proxy)

3. [Terminal User Interface (TUI)](#terminal-user-interface-tui)

   - [Navigation](#navigation)
   - [Working with Features](#working-with-features)
   - [Working with Endpoints](#working-with-endpoints)
   - [Server Management](#server-management)
   - [Dialog Navigation](#dialog-navigation)

4. [Keyboard Shortcuts](#keyboard-shortcuts)

   - [Navigation Shortcuts](#navigation-shortcuts)
   - [Feature & Endpoint Management](#feature--endpoint-management)
   - [Server & Configuration](#server--configuration)
   - [Search & Help](#search--help)
   - [Application Control](#application-control)

5. [Configuration Reference](#configuration-reference)

   - [Global Configuration](#global-configuration)
   - [Feature Configuration](#feature-configuration)
   - [Endpoint Configuration](#endpoint-configuration)
   - [Response Configuration](#response-configuration)
   - [Template Variables](#template-variables)
   - [File Structure](#file-structure)

6. [Mock Examples](#mock-examples)

   - [REST API Endpoints](#rest-api-endpoints)
   - [GraphQL Endpoint Mocking](#graphql-endpoint-mocking)
   - [Multiple Response Variations](#multiple-response-variations)

7. [Advanced Usage](#advanced-usage)

   - [Adding Response Delay](#adding-response-delay)
   - [Running in Server-Only Mode](#running-in-server-only-mode)
   - [Custom Headers](#custom-headers)
   - [Path Parameters with Regular Expressions](#path-parameters-with-regular-expressions)

8. [Troubleshooting](#troubleshooting)
   - [Server Won't Start](#server-wont-start)
   - [Changes Not Reflected](#changes-not-reflected)
   - [Proxy Not Working](#proxy-not-working)

## Getting Started

### Installation

#### Option 1: Download Binary (Recommended)

1. Download the latest release for your platform from [GitHub Releases](https://github.com/mockoho/mockoho/releases)
2. Extract the archive
3. Move the binary to a location in your PATH (optional)

#### Option 2: Using Go Install (Requires Go)

```bash
go install github.com/mockoho/mockoho@latest
```

#### Option 3: Building from Source

```bash
# Clone the repository
git clone https://github.com/mockoho/mockoho.git
cd mockoho

# Build the application
go build -o mockoho ./cmd/mockoho
```

### Running the Application

Once installed, you can run Mockoho with:

```bash
mockoho
```

This will launch the interactive CLI interface.

You can also specify a custom configuration directory:

```bash
mockoho --config /path/to/your/mocks
```

To run in server-only mode (without TUI):

```bash
mockoho server --config /path/to/your/mocks
```

### Understanding the Interface

Mockoho has a keyboard-driven interface with two main panels:

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
├────────────────────────────┴────────────────────────────────────────────────────────────────────────┤
│ t toggle  r response  o open  n new  d delete  p proxy  s server  q quit  h help                    │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Basic Usage

### Creating Your First Mock

#### 1. Create a New Feature

1. Press `Tab` to ensure you're in the Features panel
2. Press `n` to create a new feature
3. Enter a name for your feature (e.g., "products")
4. Press `Enter` to confirm

#### 2. Create a New Endpoint

1. Press `Tab` to switch to the Endpoints panel
2. Press `n` to create a new endpoint
3. Fill in the required fields:
   - **Endpoint ID**: A unique identifier (e.g., "get-products")
   - **Method**: HTTP method (e.g., "GET")
   - **Path**: API path (e.g., "/api/products")
4. Press `Enter` to confirm

#### 3. Edit the Endpoint Configuration

1. With the endpoint selected, press `o` to open the configuration file in your editor
2. Modify the JSON configuration to add more responses or customize the existing one
3. Save the file and close the editor

Example endpoint configuration:

```json
{
  "id": "get-products",
  "method": "GET",
  "path": "/api/products",
  "active": true,
  "defaultResponse": "standard",
  "responses": {
    "standard": {
      "status": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "products": [
          {
            "id": "1",
            "name": "Product 1",
            "price": 19.99
          },
          {
            "id": "2",
            "name": "Product 2",
            "price": 29.99
          }
        ],
        "total": 2
      },
      "delay": 0
    },
    "empty": {
      "status": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "products": [],
        "total": 0
      },
      "delay": 0
    },
    "error": {
      "status": 500,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      },
      "delay": 0
    }
  }
}
```

#### 4. Start the Server

1. Press `s` to start the server
2. The server will start on the configured port (default: 3000)

#### 5. Test Your Mock

You can now make requests to your mock API:

```bash
curl http://localhost:3000/api/products
```

### Working with Multiple Responses

Mockoho allows you to define multiple responses for each endpoint and easily switch between them.

#### Cycling Through Responses

1. Select an endpoint in the Endpoints panel
2. Press `r` to cycle through the available responses
3. The selected response will be used when the endpoint is called

#### Toggling Endpoints

1. Select an endpoint in the Endpoints panel
2. Press `t` to toggle the endpoint active/inactive
3. When inactive, requests to the endpoint will be proxied to the real backend (if configured)

### Configuring the Proxy

Mockoho can proxy requests to a real backend for endpoints that are not mocked or inactive.

#### Setting the Proxy Target

1. Press `p` to open the proxy configuration dialog
2. Enter the target URL (e.g., "https://api.real-server.com")
3. Press `Enter` to confirm

#### Editing the Global Configuration

You can also edit the global configuration file directly:

1. Open `mocks/config.json` in your editor
2. Modify the proxy configuration:

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

## Terminal User Interface (TUI)

### Navigation

- **Tab**: Switch between Features and Endpoints panels
- **↑/↓**: Navigate up/down in the current panel
- **Enter**: Select a feature or endpoint
- **q**: Quit the application

### Working with Features

#### Viewing Features

The Features panel shows all available features. Each feature represents a group of related endpoints.

#### Creating a New Feature

1. Press **Tab** to ensure you're in the Features panel
2. Press **n** to open the new feature dialog
3. Enter a name for your feature (e.g., "products")
4. Press **Enter** to confirm

#### Deleting a Feature

1. Select the feature you want to delete in the Features panel
2. Press **d** to open the delete confirmation dialog
3. Press **Enter** to confirm deletion

### Working with Endpoints

#### Viewing Endpoints

The Endpoints panel shows all endpoints for the selected feature. Each endpoint shows:

- HTTP method (GET, POST, PUT, etc.)
- Path (e.g., /api/users/:id)
- Active status (🟢 for active, 🔴 for inactive)
- Available responses with the default response marked with a star (★)

#### Creating a New Endpoint

1. Select a feature in the Features panel
2. Press **Tab** to switch to the Endpoints panel
3. Press **n** to open the new endpoint dialog
4. Fill in the required fields:
   - **Endpoint ID**: A unique identifier (e.g., "get-products")
   - **Method**: HTTP method (e.g., "GET")
   - **Path**: API path (e.g., "/api/products")
5. Press **Enter** to confirm

#### Toggling Endpoint Active State

1. Select an endpoint in the Endpoints panel
2. Press **t** to toggle between active (🟢) and inactive (🔴)

When an endpoint is inactive, requests to that endpoint will be proxied to the real backend (if configured).

#### Cycling Through Responses

1. Select an endpoint in the Endpoints panel
2. Press **r** to cycle through the available responses
3. The current default response is marked with a star (★)

#### Editing Endpoint Configuration

1. Select an endpoint in the Endpoints panel
2. Press **o** to open the configuration file in your default editor
3. Make changes to the JSON configuration
4. Save and close the editor
5. Press **Ctrl+r** to reload the configuration

### Server Management

#### Starting the Server

Press **s** to start the mock server. The header will update to show "Server: Running (localhost:3000)".

#### Stopping the Server

Press **s** again to stop the server. The header will update to show "Server: Stopped".

### Dialog Navigation

When in a dialog:

- Use **Tab** to navigate between input fields
- Use **Enter** to confirm
- Use **Esc** to cancel

## Keyboard Shortcuts

### Navigation Shortcuts

| Key     | Action       | Description                                  |
| ------- | ------------ | -------------------------------------------- |
| `←/→`   | Switch Panel | Switch between Features and Endpoints panels |
| `↑`     | Move Up      | Move up in the current panel                 |
| `↓`     | Move Down    | Move down in the current panel               |
| `Enter` | Select       | Select a feature or endpoint                 |

### Feature & Endpoint Management

| Key | Action   | Description                                                             |
| --- | -------- | ----------------------------------------------------------------------- |
| `n` | New      | Create new feature (in Features panel) or endpoint (in Endpoints panel) |
| `d` | Delete   | Delete selected feature or endpoint                                     |
| `t` | Toggle   | Toggle endpoint active/inactive                                         |
| `r` | Response | Cycle through available responses for the endpoint                      |

### Server & Configuration

| Key      | Action | Description                               |
| -------- | ------ | ----------------------------------------- |
| `s`      | Server | Start/stop the mock server                |
| `p`      | Proxy  | Configure proxy target                    |
| `o`      | Open   | Open configuration file in default editor |
| `Ctrl+r` | Reload | Reload configurations from disk           |

### Search & Help

| Key        | Action | Description                         |
| ---------- | ------ | ----------------------------------- |
| `/`        | Search | Search for endpoints                |
| `h` or `?` | Help   | Show help screen with all shortcuts |

### Application Control

| Key             | Action | Description          |
| --------------- | ------ | -------------------- |
| `q` or `Ctrl+c` | Quit   | Exit the application |

## Configuration Reference

### Global Configuration

The global configuration is stored in `mocks/config.json` and contains settings for the proxy server, HTTP server, and editor.

#### Proxy Configuration

The proxy configuration controls how requests are forwarded to a real backend when an endpoint is not mocked or inactive.

| Property       | Type    | Description                                                                | Default                         |
| -------------- | ------- | -------------------------------------------------------------------------- | ------------------------------- |
| `target`       | string  | The target URL to proxy requests to                                        | `"https://api.real-server.com"` |
| `changeOrigin` | boolean | Whether to change the origin of the host header to the target URL          | `true`                          |
| `pathRewrite`  | object  | A map of regex patterns to replacement strings for rewriting request paths | `{ "^/api": "" }`               |

Example:

```json
"proxyConfig": {
  "target": "https://api.real-server.com",
  "changeOrigin": true,
  "pathRewrite": {
    "^/api": ""
  }
}
```

#### Server Configuration

The server configuration controls the HTTP server settings.

| Property | Type   | Description           | Default       |
| -------- | ------ | --------------------- | ------------- |
| `port`   | number | The port to listen on | `3000`        |
| `host`   | string | The host to bind to   | `"localhost"` |

Example:

```json
"serverConfig": {
  "port": 3000,
  "host": "localhost"
}
```

#### Editor Configuration

The editor configuration controls how configuration files are opened in an external editor.

| Property  | Type   | Description                                 | Default                   |
| --------- | ------ | ------------------------------------------- | ------------------------- |
| `command` | string | The command to run to open the editor       | `"code"`                  |
| `args`    | array  | The arguments to pass to the editor command | `["-g", "{file}:{line}"]` |

Example:

```json
"editor": {
  "command": "code",
  "args": ["-g", "{file}:{line}"]
}
```

### Feature Configuration

Feature configurations are stored in separate JSON files in the `mocks` directory, with each file representing a feature (a group of related endpoints).

#### Feature Properties

| Property    | Type   | Description                         | Required |
| ----------- | ------ | ----------------------------------- | -------- |
| `feature`   | string | The name of the feature             | Yes      |
| `endpoints` | array  | An array of endpoint configurations | Yes      |

Example:

```json
{
  "feature": "users",
  "endpoints": [
    // Endpoint configurations...
  ]
}
```

### Endpoint Configuration

Each endpoint configuration defines a mock API endpoint.

| Property          | Type    | Description                                                                | Required |
| ----------------- | ------- | -------------------------------------------------------------------------- | -------- |
| `id`              | string  | A unique identifier for the endpoint                                       | Yes      |
| `method`          | string  | The HTTP method (GET, POST, PUT, DELETE, etc.)                             | Yes      |
| `path`            | string  | The URL path pattern, can include path parameters (e.g., `/api/users/:id`) | Yes      |
| `active`          | boolean | Whether the endpoint is active (if `false`, requests will be proxied)      | Yes      |
| `defaultResponse` | string  | The key of the default response to use                                     | Yes      |
| `responses`       | object  | A map of response keys to response configurations                          | Yes      |

Example:

```json
{
  "id": "get-user-profile",
  "method": "GET",
  "path": "/api/users/:id",
  "active": true,
  "defaultResponse": "standard",
  "responses": {
    // Response configurations...
  }
}
```

### Response Configuration

Each response configuration defines a possible response for an endpoint.

| Property  | Type   | Description                                               | Required |
| --------- | ------ | --------------------------------------------------------- | -------- |
| `status`  | number | The HTTP status code                                      | Yes      |
| `headers` | object | A map of header names to values                           | Yes      |
| `body`    | any    | The response body (can be an object, array, string, etc.) | Yes      |
| `delay`   | number | The delay in milliseconds before sending the response     | No       |

Example:

```json
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
}
```

### Template Variables

Mockoho supports template variables in response bodies, which are replaced with actual values when the response is generated.

#### Available Variables

| Variable          | Description                              | Example                                                                                                        |
| ----------------- | ---------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `{{params.name}}` | The value of a path parameter            | If the path is `/api/users/:id`, then `{{params.id}}` will be replaced with the actual ID from the request URL |
| `{{now}}`         | The current timestamp in ISO 8601 format | `"2023-05-13T14:30:00.000Z"`                                                                                   |

#### Usage Examples

Path parameters:

```json
{
  "id": "get-user",
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
        "name": "User {{params.id}}",
        "email": "user{{params.id}}@example.com"
      }
    }
  }
}
```

Current timestamp:

```json
{
  "id": "get-server-time",
  "method": "GET",
  "path": "/api/server/time",
  "active": true,
  "defaultResponse": "standard",
  "responses": {
    "standard": {
      "status": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "timestamp": "{{now}}",
        "timezone": "UTC"
      }
    }
  }
}
```

### File Structure

#### Directory Layout

Mockoho uses the following directory structure:

```
mocks/
  ├── config.json     # Global configuration
  ├── users.json      # Feature configuration for users
  ├── products.json   # Feature configuration for products
  └── ...             # Other feature configurations
```

#### File Naming Conventions

- The global configuration file must be named `config.json`.
- Feature configuration files should be named after the feature they represent, with a `.json` extension.
- File names should be lowercase and use hyphens for spaces (e.g., `user-profiles.json`).

## Mock Examples

### REST API Endpoints

#### GET Request with Query Parameters

```json
{
  "feature": "products",
  "endpoints": [
    {
      "id": "get-products-list",
      "method": "GET",
      "path": "/api/products",
      "active": true,
      "defaultResponse": "standard",
      "responses": {
        "standard": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "products": [
              {
                "id": "1",
                "name": "Product 1",
                "price": 19.99,
                "category": "electronics"
              },
              {
                "id": "2",
                "name": "Product 2",
                "price": 29.99,
                "category": "clothing"
              },
              {
                "id": "3",
                "name": "Product 3",
                "price": 39.99,
                "category": "electronics"
              }
            ],
            "total": 3,
            "page": 1,
            "pageSize": 10
          },
          "delay": 0
        },
        "filtered": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "products": [
              {
                "id": "1",
                "name": "Product 1",
                "price": 19.99,
                "category": "electronics"
              },
              {
                "id": "3",
                "name": "Product 3",
                "price": 39.99,
                "category": "electronics"
              }
            ],
            "total": 2,
            "page": 1,
            "pageSize": 10
          },
          "delay": 0
        },
        "empty": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "products": [],
            "total": 0,
            "page": 1,
            "pageSize": 10
          },
          "delay": 0
        }
      }
    }
  ]
}
```

#### POST Request with Request Body

```json
{
  "feature": "products",
  "endpoints": [
    {
      "id": "create-product",
      "method": "POST",
      "path": "/api/products",
      "active": true,
      "defaultResponse": "success",
      "responses": {
        "success": {
          "status": 201,
          "headers": {
            "Content-Type": "application/json",
            "Location": "/api/products/4"
          },
          "body": {
            "id": "4",
            "name": "New Product",
            "price": 49.99,
            "category": "home",
            "createdAt": "{{now}}"
          },
          "delay": 0
        },
        "validation-error": {
          "status": 400,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "Validation failed",
            "fields": {
              "price": "Price must be a positive number"
            }
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### GraphQL Endpoint Mocking

#### Basic GraphQL Query

```json
{
  "feature": "graphql",
  "endpoints": [
    {
      "id": "graphql-query",
      "method": "POST",
      "path": "/api/graphql",
      "active": true,
      "defaultResponse": "products-query",
      "responses": {
        "products-query": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "data": {
              "products": [
                {
                  "id": "1",
                  "name": "Product 1",
                  "price": 19.99,
                  "category": {
                    "id": "cat1",
                    "name": "Electronics"
                  }
                },
                {
                  "id": "2",
                  "name": "Product 2",
                  "price": 29.99,
                  "category": {
                    "id": "cat2",
                    "name": "Clothing"
                  }
                }
              ]
            }
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### Multiple Response Variations

#### Success/Error Scenarios

```json
{
  "feature": "auth",
  "endpoints": [
    {
      "id": "login",
      "method": "POST",
      "path": "/api/auth/login",
      "active": true,
      "defaultResponse": "success",
      "responses": {
        "success": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "user": {
              "id": "1",
              "name": "John Doe",
              "email": "john@example.com"
            }
          },
          "delay": 0
        },
        "invalid-credentials": {
          "status": 401,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "Invalid credentials",
            "code": "INVALID_CREDENTIALS"
          },
          "delay": 0
        },
        "account-locked": {
          "status": 403,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "Account locked",
            "code": "ACCOUNT_LOCKED"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

## Advanced Usage

### Adding Response Delay

You can simulate network latency by adding a delay to your responses:

```json
"responses": {
  "slow": {
    "status": 200,
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "message": "This response was delayed"
    },
    "delay": 2000  // 2 seconds delay
  }
}
```

### Running in Server-Only Mode

To run Mockoho without the UI (useful for CI/CD environments):

```bash
mockoho server
```

### Custom Headers

You can add custom headers to your responses:

```json
"headers": {
  "Content-Type": "application/json",
  "X-Custom-Header": "Custom Value",
  "X-Rate-Limit": "100",
  "X-Rate-Limit-Remaining": "99"
}
```

### Path Parameters with Regular Expressions

Path parameters support simple patterns:

```json
{
  "id": "get-user-files",
  "method": "GET",
  "path": "/api/users/:userId/files/:fileId",
  "active": true,
  "defaultResponse": "standard",
  "responses": {
    "standard": {
      "status": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "userId": "{{params.userId}}",
        "fileId": "{{params.fileId}}",
        "fileName": "File {{params.fileId}}"
      }
    }
  }
}
```

## Troubleshooting

### Server Won't Start

- Check if the port is already in use
- Verify that the configuration files are valid JSON

### Changes Not Reflected

- Press `Ctrl+r` to reload the configuration
- Check for JSON syntax errors in your configuration files

### Proxy Not Working

- Verify that the proxy target is correct and accessible
- Check that the endpoint is inactive or not defined
