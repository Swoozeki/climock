# Mockoho Configuration Reference

This document provides a comprehensive reference for all configuration options available in Mockoho.

## Table of Contents

1. [Global Configuration](#global-configuration)

   - [Proxy Configuration](#proxy-configuration)
   - [Server Configuration](#server-configuration)
   - [Editor Configuration](#editor-configuration)

2. [Feature Configuration](#feature-configuration)

   - [Feature Properties](#feature-properties)
   - [Endpoint Configuration](#endpoint-configuration)
   - [Response Configuration](#response-configuration)

3. [Template Variables](#template-variables)

   - [Available Variables](#available-variables)
   - [Usage Examples](#usage-examples)

4. [File Structure](#file-structure)
   - [Directory Layout](#directory-layout)
   - [File Naming Conventions](#file-naming-conventions)

## Global Configuration

The global configuration is stored in `mocks/config.json` and contains settings for the proxy server, HTTP server, and editor.

### Proxy Configuration

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

### Server Configuration

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

### Editor Configuration

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

## Feature Configuration

Feature configurations are stored in separate JSON files in the `mocks` directory, with each file representing a feature (a group of related endpoints).

### Feature Properties

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

## Template Variables

Mockoho supports template variables in response bodies, which are replaced with actual values when the response is generated.

### Available Variables

| Variable          | Description                              | Example                                                                                                        |
| ----------------- | ---------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `{{params.name}}` | The value of a path parameter            | If the path is `/api/users/:id`, then `{{params.id}}` will be replaced with the actual ID from the request URL |
| `{{now}}`         | The current timestamp in ISO 8601 format | `"2023-05-13T14:30:00.000Z"`                                                                                   |

### Usage Examples

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

## File Structure

### Directory Layout

Mockoho uses the following directory structure:

```
mocks/
  ├── config.json     # Global configuration
  ├── users.json      # Feature configuration for users
  ├── products.json   # Feature configuration for products
  └── ...             # Other feature configurations
```

### File Naming Conventions

- The global configuration file must be named `config.json`.
- Feature configuration files should be named after the feature they represent, with a `.json` extension.
- File names should be lowercase and use hyphens for spaces (e.g., `user-profiles.json`).

## Advanced Configuration

### Response Delay

You can simulate network latency by adding a delay to your responses:

```json
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

### Binary Responses

For binary responses, you can use a base64-encoded string:

```json
"image": {
  "status": 200,
  "headers": {
    "Content-Type": "image/png"
  },
  "body": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+P+/HgAFhAJ/wlseKgAAAABJRU5ErkJggg==",
  "delay": 0
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

### Multiple Response Variations

You can define multiple responses for an endpoint and switch between them using the `r` key in the UI:

```json
"responses": {
  "success": {
    "status": 200,
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "message": "Success"
    }
  },
  "error": {
    "status": 500,
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "error": "Internal server error"
    }
  },
  "not-found": {
    "status": 404,
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "error": "Not found"
    }
  }
}
```

## Best Practices

1. **Organize by Feature**: Group related endpoints into feature files for better organization.

2. **Use Descriptive IDs**: Give endpoints descriptive IDs that indicate their purpose.

3. **Use Path Parameters**: Use path parameters (`:id`) instead of hardcoding values in paths.

4. **Provide Multiple Responses**: Define multiple responses for each endpoint to simulate different scenarios.

5. **Use Template Variables**: Use template variables to make responses dynamic.

6. **Add Delays for Realism**: Add delays to responses to simulate network latency.

7. **Keep Configuration Files Small**: Split large feature files into smaller ones for better maintainability.

8. **Use Consistent Naming**: Use consistent naming conventions for files, features, endpoints, and responses.

9. **Document Your Mocks**: Add comments or documentation to explain the purpose of each endpoint and response.

10. **Version Your Mocks**: Use version control to track changes to your mock configurations.
