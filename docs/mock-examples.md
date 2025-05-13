# Mockoho Mock Configuration Examples

This document provides example mock configurations for common use cases to help you get started with Mockoho.

## Table of Contents

1. [REST API Endpoints](#rest-api-endpoints)

   - [GET Request with Query Parameters](#get-request-with-query-parameters)
   - [POST Request with Request Body](#post-request-with-request-body)
   - [PUT Request for Updates](#put-request-for-updates)
   - [DELETE Request](#delete-request)
   - [PATCH Request for Partial Updates](#patch-request-for-partial-updates)

2. [GraphQL Endpoint Mocking](#graphql-endpoint-mocking)

   - [Basic GraphQL Query](#basic-graphql-query)
   - [GraphQL Mutation](#graphql-mutation)

3. [Multiple Response Variations](#multiple-response-variations)

   - [Success/Error Scenarios](#successerror-scenarios)
   - [Different Data States](#different-data-states)
   - [Pagination Examples](#pagination-examples)

4. [Template Variables](#template-variables)
   - [Path Parameters](#path-parameters)
   - [Current Timestamp](#current-timestamp)
   - [Complex Template Usage](#complex-template-usage)

## REST API Endpoints

### GET Request with Query Parameters

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

### POST Request with Request Body

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

### PUT Request for Updates

```json
{
  "feature": "products",
  "endpoints": [
    {
      "id": "update-product",
      "method": "PUT",
      "path": "/api/products/:id",
      "active": true,
      "defaultResponse": "success",
      "responses": {
        "success": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "{{params.id}}",
            "name": "Updated Product",
            "price": 59.99,
            "category": "electronics",
            "updatedAt": "{{now}}"
          },
          "delay": 0
        },
        "not-found": {
          "status": 404,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "Product not found",
            "code": "PRODUCT_NOT_FOUND"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### DELETE Request

```json
{
  "feature": "products",
  "endpoints": [
    {
      "id": "delete-product",
      "method": "DELETE",
      "path": "/api/products/:id",
      "active": true,
      "defaultResponse": "success",
      "responses": {
        "success": {
          "status": 204,
          "headers": {},
          "body": "",
          "delay": 0
        },
        "not-found": {
          "status": 404,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "Product not found",
            "code": "PRODUCT_NOT_FOUND"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### PATCH Request for Partial Updates

```json
{
  "feature": "products",
  "endpoints": [
    {
      "id": "patch-product",
      "method": "PATCH",
      "path": "/api/products/:id",
      "active": true,
      "defaultResponse": "success",
      "responses": {
        "success": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "{{params.id}}",
            "price": 69.99,
            "updatedAt": "{{now}}"
          },
          "delay": 0
        },
        "not-found": {
          "status": 404,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "error": "Product not found",
            "code": "PRODUCT_NOT_FOUND"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

## GraphQL Endpoint Mocking

### Basic GraphQL Query

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
        },
        "product-detail-query": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "data": {
              "product": {
                "id": "1",
                "name": "Product 1",
                "description": "Detailed product description",
                "price": 19.99,
                "category": {
                  "id": "cat1",
                  "name": "Electronics"
                },
                "reviews": [
                  {
                    "id": "rev1",
                    "rating": 4,
                    "comment": "Great product!"
                  },
                  {
                    "id": "rev2",
                    "rating": 5,
                    "comment": "Excellent quality!"
                  }
                ]
              }
            }
          },
          "delay": 0
        },
        "error-response": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "errors": [
              {
                "message": "Field 'product' argument 'id' of type 'ID!' is required, but it was not provided.",
                "locations": [
                  {
                    "line": 2,
                    "column": 3
                  }
                ],
                "path": ["product"]
              }
            ]
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### GraphQL Mutation

```json
{
  "feature": "graphql",
  "endpoints": [
    {
      "id": "graphql-mutation",
      "method": "POST",
      "path": "/api/graphql",
      "active": true,
      "defaultResponse": "create-product",
      "responses": {
        "create-product": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "data": {
              "createProduct": {
                "id": "3",
                "name": "New Product",
                "price": 49.99,
                "category": {
                  "id": "cat1",
                  "name": "Electronics"
                }
              }
            }
          },
          "delay": 0
        },
        "update-product": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "data": {
              "updateProduct": {
                "id": "1",
                "name": "Updated Product Name",
                "price": 24.99,
                "category": {
                  "id": "cat1",
                  "name": "Electronics"
                }
              }
            }
          },
          "delay": 0
        },
        "delete-product": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "data": {
              "deleteProduct": {
                "id": "2",
                "success": true
              }
            }
          },
          "delay": 0
        },
        "validation-error": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "errors": [
              {
                "message": "Price must be a positive number",
                "path": ["createProduct", "price"],
                "extensions": {
                  "code": "VALIDATION_ERROR",
                  "field": "price"
                }
              }
            ],
            "data": null
          },
          "delay": 0
        }
      }
    }
  ]
}
```

## Multiple Response Variations

### Success/Error Scenarios

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
              "id": "user123",
              "name": "John Doe",
              "email": "john@example.com"
            },
            "expiresAt": "{{now}}"
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
            "error": "Account locked due to too many failed attempts",
            "code": "ACCOUNT_LOCKED",
            "unlockAt": "{{now}}"
          },
          "delay": 0
        },
        "server-error": {
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
  ]
}
```

### Different Data States

```json
{
  "feature": "cart",
  "endpoints": [
    {
      "id": "get-cart",
      "method": "GET",
      "path": "/api/cart",
      "active": true,
      "defaultResponse": "standard",
      "responses": {
        "empty": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "items": [],
            "total": 0,
            "currency": "USD"
          },
          "delay": 0
        },
        "standard": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "items": [
              {
                "id": "item1",
                "productId": "prod1",
                "name": "Product 1",
                "quantity": 2,
                "price": 19.99,
                "total": 39.98
              }
            ],
            "total": 39.98,
            "currency": "USD"
          },
          "delay": 0
        },
        "multiple-items": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "items": [
              {
                "id": "item1",
                "productId": "prod1",
                "name": "Product 1",
                "quantity": 2,
                "price": 19.99,
                "total": 39.98
              },
              {
                "id": "item2",
                "productId": "prod2",
                "name": "Product 2",
                "quantity": 1,
                "price": 29.99,
                "total": 29.99
              },
              {
                "id": "item3",
                "productId": "prod3",
                "name": "Product 3",
                "quantity": 3,
                "price": 9.99,
                "total": 29.97
              }
            ],
            "total": 99.94,
            "currency": "USD"
          },
          "delay": 0
        },
        "with-discount": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "items": [
              {
                "id": "item1",
                "productId": "prod1",
                "name": "Product 1",
                "quantity": 2,
                "price": 19.99,
                "total": 39.98
              },
              {
                "id": "item2",
                "productId": "prod2",
                "name": "Product 2",
                "quantity": 1,
                "price": 29.99,
                "total": 29.99
              }
            ],
            "subtotal": 69.97,
            "discount": {
              "code": "SUMMER20",
              "amount": 14.0
            },
            "total": 55.97,
            "currency": "USD"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### Pagination Examples

```json
{
  "feature": "products",
  "endpoints": [
    {
      "id": "get-products-paginated",
      "method": "GET",
      "path": "/api/products/paginated",
      "active": true,
      "defaultResponse": "page1",
      "responses": {
        "page1": {
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
              },
              {
                "id": "3",
                "name": "Product 3",
                "price": 39.99
              }
            ],
            "pagination": {
              "page": 1,
              "pageSize": 3,
              "totalItems": 9,
              "totalPages": 3,
              "hasNext": true,
              "hasPrev": false
            }
          },
          "delay": 0
        },
        "page2": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "products": [
              {
                "id": "4",
                "name": "Product 4",
                "price": 49.99
              },
              {
                "id": "5",
                "name": "Product 5",
                "price": 59.99
              },
              {
                "id": "6",
                "name": "Product 6",
                "price": 69.99
              }
            ],
            "pagination": {
              "page": 2,
              "pageSize": 3,
              "totalItems": 9,
              "totalPages": 3,
              "hasNext": true,
              "hasPrev": true
            }
          },
          "delay": 0
        },
        "page3": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "products": [
              {
                "id": "7",
                "name": "Product 7",
                "price": 79.99
              },
              {
                "id": "8",
                "name": "Product 8",
                "price": 89.99
              },
              {
                "id": "9",
                "name": "Product 9",
                "price": 99.99
              }
            ],
            "pagination": {
              "page": 3,
              "pageSize": 3,
              "totalItems": 9,
              "totalPages": 3,
              "hasNext": false,
              "hasPrev": true
            }
          },
          "delay": 0
        },
        "empty-page": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "products": [],
            "pagination": {
              "page": 4,
              "pageSize": 3,
              "totalItems": 9,
              "totalPages": 3,
              "hasNext": false,
              "hasPrev": true
            }
          },
          "delay": 0
        }
      }
    }
  ]
}
```

## Template Variables

### Path Parameters

```json
{
  "feature": "users",
  "endpoints": [
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
            "email": "user{{params.id}}@example.com",
            "createdAt": "2023-01-01T00:00:00Z"
          },
          "delay": 0
        }
      }
    },
    {
      "id": "get-user-post",
      "method": "GET",
      "path": "/api/users/:userId/posts/:postId",
      "active": true,
      "defaultResponse": "standard",
      "responses": {
        "standard": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "{{params.postId}}",
            "userId": "{{params.userId}}",
            "title": "Post {{params.postId}} by User {{params.userId}}",
            "content": "This is the content of post {{params.postId}} by user {{params.userId}}",
            "createdAt": "2023-01-01T00:00:00Z"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### Current Timestamp

```json
{
  "feature": "events",
  "endpoints": [
    {
      "id": "create-event",
      "method": "POST",
      "path": "/api/events",
      "active": true,
      "defaultResponse": "success",
      "responses": {
        "success": {
          "status": 201,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "event123",
            "title": "New Event",
            "description": "Event description",
            "createdAt": "{{now}}",
            "updatedAt": "{{now}}"
          },
          "delay": 0
        }
      }
    },
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
          },
          "delay": 0
        }
      }
    }
  ]
}
```

### Complex Template Usage

```json
{
  "feature": "orders",
  "endpoints": [
    {
      "id": "get-order",
      "method": "GET",
      "path": "/api/orders/:id",
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
            "customer": {
              "id": "cust123",
              "name": "John Doe"
            },
            "items": [
              {
                "id": "item1",
                "productId": "prod1",
                "name": "Product 1",
                "quantity": 2,
                "price": 19.99,
                "total": 39.98
              },
              {
                "id": "item2",
                "productId": "prod2",
                "name": "Product 2",
                "quantity": 1,
                "price": 29.99,
                "total": 29.99
              }
            ],
            "shipping": {
              "address": "123 Main St",
              "city": "Anytown",
              "state": "CA",
              "zipCode": "12345",
              "country": "USA"
            },
            "payment": {
              "method": "credit_card",
              "cardLast4": "1234",
              "status": "paid"
            },
            "subtotal": 69.97,
            "tax": 5.6,
            "shipping": 4.99,
            "total": 80.56,
            "currency": "USD",
            "status": "shipped",
            "createdAt": "2023-01-01T00:00:00Z",
            "updatedAt": "{{now}}"
          },
          "delay": 0
        },
        "processing": {
          "status": 200,
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "id": "{{params.id}}",
            "customer": {
              "id": "cust123",
              "name": "John Doe"
            },
            "items": [
              {
                "id": "item1",
                "productId": "prod1",
                "name": "Product 1",
                "quantity": 2,
                "price": 19.99,
                "total": 39.98
              },
              {
                "id": "item2",
                "productId": "prod2",
                "name": "Product 2",
                "quantity": 1,
                "price": 29.99,
                "total": 29.99
              }
            ],
            "shipping": {
              "address": "123 Main St",
              "city": "Anytown",
              "state": "CA",
              "zipCode": "12345",
              "country": "USA"
            },
            "payment": {
              "method": "credit_card",
              "cardLast4": "1234",
              "status": "processing"
            },
            "subtotal": 69.97,
            "tax": 5.6,
            "shipping": 4.99,
            "total": 80.56,
            "currency": "USD",
            "status": "processing",
            "createdAt": "2023-01-01T00:00:00Z",
            "updatedAt": "{{now}}"
          },
          "delay": 0
        }
      }
    }
  ]
}
```

These examples cover a wide range of common use cases for mocking APIs with Mockoho. You can use them as a starting point for your own mock configurations, adapting them to your specific needs.
