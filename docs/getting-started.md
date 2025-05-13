# Getting Started with Mockoho

This guide will help you get up and running with Mockoho, a powerful mock server system designed to facilitate frontend development without a ready backend and to easily update responses for demo purposes.

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

### Running the Application

Once built, you can run Mockoho with:

```bash
./mockoho
```

This will launch the interactive CLI interface.

## Understanding the Interface

Mockoho has a keyboard-driven interface with two main panels:

1. **Features Panel** (left): Lists all available features (groups of endpoints)
2. **Endpoints Panel** (right): Lists all endpoints for the selected feature

### Navigation

- Use `Tab` to switch between the Features and Endpoints panels
- Use `↑/↓` arrow keys to navigate within a panel
- Use `Enter` to select an item

## Creating Your First Mock

### 1. Create a New Feature

1. Press `Tab` to ensure you're in the Features panel
2. Press `n` to create a new feature
3. Enter a name for your feature (e.g., "products")
4. Press `Enter` to confirm

### 2. Create a New Endpoint

1. Press `Tab` to switch to the Endpoints panel
2. Press `n` to create a new endpoint
3. Fill in the required fields:
   - **Endpoint ID**: A unique identifier (e.g., "get-products")
   - **Method**: HTTP method (e.g., "GET")
   - **Path**: API path (e.g., "/api/products")
4. Press `Enter` to confirm

### 3. Edit the Endpoint Configuration

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

### 4. Start the Server

1. Press `s` to start the server
2. The server will start on the configured port (default: 3000)

### 5. Test Your Mock

You can now make requests to your mock API:

```bash
curl http://localhost:3000/api/products
```

## Working with Multiple Responses

Mockoho allows you to define multiple responses for each endpoint and easily switch between them.

### Cycling Through Responses

1. Select an endpoint in the Endpoints panel
2. Press `r` to cycle through the available responses
3. The selected response will be used when the endpoint is called

### Toggling Endpoints

1. Select an endpoint in the Endpoints panel
2. Press `t` to toggle the endpoint active/inactive
3. When inactive, requests to the endpoint will be proxied to the real backend (if configured)

## Configuring the Proxy

Mockoho can proxy requests to a real backend for endpoints that are not mocked or inactive.

### Setting the Proxy Target

1. Press `p` to open the proxy configuration dialog
2. Enter the target URL (e.g., "https://api.real-server.com")
3. Press `Enter` to confirm

### Editing the Global Configuration

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

## Using Template Variables

Mockoho supports template variables in response bodies:

- `{{params.id}}` - Path parameter value (e.g., `:id` in `/api/users/:id`)
- `{{now}}` - Current timestamp in ISO 8601 format

Example:

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
        "name": "John Doe",
        "lastAccessed": "{{now}}"
      },
      "delay": 0
    }
  }
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
./mockoho server
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

## Next Steps

- Explore the [Configuration Reference](configuration-reference.md) for detailed information on all configuration options
- Check out the [Mock Examples](mock-examples.md) for common API patterns and how to mock them
- Learn about the [TUI Architecture](tui-architecture.md) for more details on the terminal UI
