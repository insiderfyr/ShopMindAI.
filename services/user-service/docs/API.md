# User Service API Documentation

## Overview

The User Service is responsible for managing user accounts, authentication, and profiles in the ShopGPT platform. It provides RESTful APIs for user registration, authentication, profile management, and user discovery.

## Base URL

```
https://api.shopgpt.com/v1/users
```

## Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

## Error Responses

All endpoints return errors in the following format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

Common HTTP status codes:
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

## Endpoints

### 1. Create User (Register)

Create a new user account.

**Endpoint:** `POST /users`

**Request Body:**
```json
{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "SecurePassword123!",
  "fullName": "John Doe"
}
```

**Validation Rules:**
- Email: Valid email format, required
- Username: 3-20 characters, alphanumeric and underscore only
- Password: Minimum 8 characters, must contain uppercase, lowercase, number, and special character
- Full Name: 2-50 characters

**Response:** `201 Created`
```json
{
  "id": "user-123-uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "fullName": "John Doe",
  "avatar": null,
  "bio": null,
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400` - Invalid request body or validation failure
- `409` - User with email already exists

**Example cURL:**
```bash
curl -X POST https://api.shopgpt.com/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "SecurePassword123!",
    "fullName": "John Doe"
  }'
```

### 2. Get User by ID

Retrieve a specific user's public profile.

**Endpoint:** `GET /users/{id}`

**Path Parameters:**
- `id` - User ID (UUID)

**Response:** `200 OK`
```json
{
  "id": "user-123-uuid",
  "username": "johndoe",
  "fullName": "John Doe",
  "avatar": "https://cdn.shopgpt.com/avatars/user-123.jpg",
  "bio": "Shopping enthusiast",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `404` - User not found

**Example cURL:**
```bash
curl -X GET https://api.shopgpt.com/v1/users/user-123-uuid
```

### 3. Update User Profile

Update the authenticated user's profile.

**Endpoint:** `PUT /users/{id}`

**Authentication:** Required (can only update own profile)

**Path Parameters:**
- `id` - User ID (must match authenticated user)

**Request Body:**
```json
{
  "username": "newusername",
  "fullName": "John Smith",
  "bio": "Love finding great deals!",
  "avatar": "https://example.com/avatar.jpg"
}
```

**Response:** `200 OK`
```json
{
  "id": "user-123-uuid",
  "email": "user@example.com",
  "username": "newusername",
  "fullName": "John Smith",
  "avatar": "https://example.com/avatar.jpg",
  "bio": "Love finding great deals!",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T14:20:00Z"
}
```

**Error Responses:**
- `401` - Unauthorized
- `403` - Forbidden (trying to update another user)
- `404` - User not found

**Example cURL:**
```bash
curl -X PUT https://api.shopgpt.com/v1/users/user-123-uuid \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newusername",
    "bio": "Love finding great deals!"
  }'
```

### 4. Delete User

Soft delete a user account.

**Endpoint:** `DELETE /users/{id}`

**Authentication:** Required (can only delete own account)

**Path Parameters:**
- `id` - User ID (must match authenticated user)

**Response:** `204 No Content`

**Error Responses:**
- `401` - Unauthorized
- `403` - Forbidden (trying to delete another user)

**Example cURL:**
```bash
curl -X DELETE https://api.shopgpt.com/v1/users/user-123-uuid \
  -H "Authorization: Bearer <jwt_token>"
```

### 5. List Users

Get a paginated list of users.

**Endpoint:** `GET /users`

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)
- `search` - Search by username or full name

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": "user-123-uuid",
      "username": "johndoe",
      "fullName": "John Doe",
      "avatar": "https://cdn.shopgpt.com/avatars/user-123.jpg",
      "createdAt": "2024-01-15T10:30:00Z"
    },
    {
      "id": "user-456-uuid",
      "username": "janesmith",
      "fullName": "Jane Smith",
      "avatar": null,
      "createdAt": "2024-01-14T09:15:00Z"
    }
  ],
  "total": 150,
  "page": "1",
  "limit": "20",
  "hasNext": true,
  "hasPrev": false
}
```

**Example cURL:**
```bash
curl -X GET "https://api.shopgpt.com/v1/users?page=1&limit=20&search=john"
```

### 6. Get Current User Profile

Get the authenticated user's complete profile.

**Endpoint:** `GET /users/me`

**Authentication:** Required

**Response:** `200 OK`
```json
{
  "id": "user-123-uuid",
  "email": "user@example.com",
  "username": "johndoe",
  "fullName": "John Doe",
  "avatar": "https://cdn.shopgpt.com/avatars/user-123.jpg",
  "bio": "Shopping enthusiast",
  "preferences": {
    "favoriteStores": ["amazon", "bestbuy"],
    "categories": ["electronics", "books"],
    "priceAlerts": true,
    "emailNotifications": true
  },
  "stats": {
    "searchesCount": 245,
    "savedProducts": 67,
    "priceAlertsActive": 12
  },
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T14:20:00Z"
}
```

**Error Responses:**
- `401` - Unauthorized

**Example cURL:**
```bash
curl -X GET https://api.shopgpt.com/v1/users/me \
  -H "Authorization: Bearer <jwt_token>"
```

### 7. Update Password

Change the authenticated user's password.

**Endpoint:** `PUT /users/password`

**Authentication:** Required

**Request Body:**
```json
{
  "oldPassword": "CurrentPassword123!",
  "newPassword": "NewSecurePassword456!"
}
```

**Validation Rules:**
- New password must meet security requirements
- Cannot reuse the last 5 passwords

**Response:** `200 OK`
```json
{
  "message": "Password updated successfully"
}
```

**Error Responses:**
- `400` - Invalid old password or new password doesn't meet requirements
- `401` - Unauthorized

**Example cURL:**
```bash
curl -X PUT https://api.shopgpt.com/v1/users/password \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "oldPassword": "CurrentPassword123!",
    "newPassword": "NewSecurePassword456!"
  }'
```

### 8. Upload Avatar

Upload a new avatar image.

**Endpoint:** `POST /users/avatar`

**Authentication:** Required

**Request:** Multipart form data
- `avatar` - Image file (JPEG, PNG, max 5MB)

**Response:** `200 OK`
```json
{
  "avatarUrl": "https://cdn.shopgpt.com/avatars/user-123-uuid.jpg"
}
```

**Error Responses:**
- `400` - Invalid file format or size
- `401` - Unauthorized

**Example cURL:**
```bash
curl -X POST https://api.shopgpt.com/v1/users/avatar \
  -H "Authorization: Bearer <jwt_token>" \
  -F "avatar=@/path/to/image.jpg"
```

### 9. Get User Preferences

Get user preferences and settings.

**Endpoint:** `GET /users/{id}/preferences`

**Authentication:** Required (can only access own preferences)

**Response:** `200 OK`
```json
{
  "userId": "user-123-uuid",
  "favoriteStores": ["amazon", "bestbuy", "walmart"],
  "categories": ["electronics", "home", "books"],
  "priceAlerts": {
    "enabled": true,
    "threshold": 10,
    "frequency": "daily"
  },
  "notifications": {
    "email": true,
    "push": true,
    "sms": false,
    "newsletter": true
  },
  "privacy": {
    "profileVisible": true,
    "showActivity": false
  },
  "language": "en",
  "currency": "USD",
  "timezone": "America/New_York"
}
```

### 10. Update User Preferences

Update user preferences and settings.

**Endpoint:** `PUT /users/{id}/preferences`

**Authentication:** Required (can only update own preferences)

**Request Body:**
```json
{
  "favoriteStores": ["amazon", "target"],
  "priceAlerts": {
    "enabled": true,
    "threshold": 15
  },
  "notifications": {
    "email": false,
    "push": true
  }
}
```

**Response:** `200 OK`
```json
{
  "message": "Preferences updated successfully",
  "preferences": {
    // Updated preferences object
  }
}
```

## Rate Limiting

API calls are rate limited per user:
- Authenticated users: 1000 requests per hour
- Unauthenticated users: 100 requests per hour

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642255200
```

## Webhooks

The User Service can send webhooks for the following events:
- `user.created`
- `user.updated`
- `user.deleted`
- `user.password_changed`

## SDKs

Official SDKs are available for:
- JavaScript/TypeScript
- Python
- Go
- Java

## Postman Collection

Download our [Postman Collection](https://api.shopgpt.com/docs/postman/user-service.json) for easy API testing.

## OpenAPI Specification

View the complete [OpenAPI 3.0 specification](https://api.shopgpt.com/docs/openapi/user-service.yaml).

## Support

For API support, please contact:
- Email: api-support@shopgpt.com
- Developer Portal: https://developers.shopgpt.com
- Status Page: https://status.shopgpt.com