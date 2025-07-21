# ShopMindAI API Documentation

## Base URLs

- **Development:** `http://localhost:8080`
- **Production:** `https://api.shopmindai.com`
- **WebSocket:** `ws://localhost:8082/ws` (Chat Service)

## Authentication

All API requests (except `/auth/*` endpoints) require JWT authentication:

```http
Authorization: Bearer <JWT_TOKEN>
```

## API Endpoints

### 🔐 Authentication Service

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}

Response 201:
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-20T15:30:00Z"
}

Response 400:
{
  "error": "Email already exists"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response 200:
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe"
  }
}

Response 401:
{
  "error": "Invalid credentials"
}
```

### 👤 User Service

#### Create User Profile
```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "preferences": {
    "categories": ["electronics", "fashion", "home"],
    "budget": {
      "min": 100,
      "max": 5000,
      "currency": "USD"
    },
    "brands": ["Apple", "Samsung", "Nike"],
    "notifications": {
      "email": true,
      "push": true,
      "sms": false
    }
  }
}

Response 201:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "preferences": {...},
  "created_at": "2024-01-20T15:30:00Z",
  "updated_at": "2024-01-20T15:30:00Z"
}
```

#### Get User Profile
```http
GET /api/v1/users/{user_id}
Authorization: Bearer <token>

Response 200:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "preferences": {
    "categories": ["electronics", "fashion"],
    "budget": {
      "min": 100,
      "max": 5000,
      "currency": "USD"
    },
    "brands": ["Apple", "Samsung"],
    "notifications": {
      "email": true,
      "push": true,
      "sms": false
    }
  },
  "stats": {
    "total_searches": 145,
    "products_viewed": 523,
    "products_purchased": 12,
    "money_saved": 234.50
  },
  "created_at": "2024-01-20T15:30:00Z",
  "updated_at": "2024-01-25T10:15:00Z"
}
```

### 💬 Chat Service

#### Create Chat Session
```http
POST /api/v1/chat/sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Shopping for a new laptop",
  "context": {
    "budget": 1500,
    "purpose": "gaming and work"
  }
}

Response 201:
{
  "session_id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Shopping for a new laptop",
  "created_at": "2024-01-20T15:30:00Z"
}
```

#### Send Message
```http
POST /api/v1/chat/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "session_id": "660e8400-e29b-41d4-a716-446655440001",
  "content": "Find me a gaming laptop under $1500 with RTX 4060",
  "attachments": []
}

Response 200:
{
  "message_id": "770e8400-e29b-41d4-a716-446655440002",
  "session_id": "660e8400-e29b-41d4-a716-446655440001",
  "role": "user",
  "content": "Find me a gaming laptop under $1500 with RTX 4060",
  "timestamp": "2024-01-20T15:31:00Z",
  "ai_response": {
    "message_id": "770e8400-e29b-41d4-a716-446655440003",
    "content": "I found 3 excellent gaming laptops that match your criteria...",
    "recommendations": [
      {
        "product_id": "ASUS-ROG-G15",
        "name": "ASUS ROG Strix G15",
        "price": 1299.99,
        "original_price": 1499.99,
        "discount": "13%",
        "specs": {
          "gpu": "RTX 4060",
          "cpu": "AMD Ryzen 7 7735HS",
          "ram": "16GB DDR5",
          "storage": "512GB NVMe SSD"
        },
        "rating": 4.5,
        "reviews": 234,
        "match_score": 0.95,
        "retailers": [
          {
            "name": "Amazon",
            "price": 1299.99,
            "in_stock": true,
            "delivery": "2-day shipping",
            "url": "https://amazon.com/..."
          },
          {
            "name": "Best Buy",
            "price": 1349.99,
            "in_stock": true,
            "delivery": "In-store pickup available",
            "url": "https://bestbuy.com/..."
          }
        ]
      }
    ],
    "analysis": {
      "pros": [
        "Excellent performance for gaming and work",
        "RTX 4060 can handle modern games at 1080p/1440p",
        "Good cooling system"
      ],
      "cons": [
        "Battery life is average (4-5 hours)",
        "Can get loud under heavy load"
      ],
      "best_deal": {
        "retailer": "Amazon",
        "savings": 200.00,
        "reason": "Lowest price + fast shipping"
      }
    }
  }
}
```

#### WebSocket Connection (Real-time Chat)
```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8082/ws?session_id=660e8400-e29b-41d4-a716-446655440001');

// Send message
ws.send(JSON.stringify({
  type: 'message',
  content: 'Show me more options with better battery life'
}));

// Receive streaming response
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch(data.type) {
    case 'stream':
      // Partial response (real-time typing effect)
      console.log('AI typing:', data.chunk);
      break;
      
    case 'recommendation':
      // Product recommendation
      console.log('New product:', data.product);
      break;
      
    case 'complete':
      // Full response ready
      console.log('AI response complete:', data.content);
      break;
      
    case 'price_alert':
      // Real-time price drop notification
      console.log('Price dropped!', data.product, data.new_price);
      break;
  }
};
```

### 🔍 Search & Recommendations

#### Search Products
```http
POST /api/v1/search
Authorization: Bearer <token>
Content-Type: application/json

{
  "query": "wireless headphones noise cancelling",
  "filters": {
    "price_range": {
      "min": 100,
      "max": 400
    },
    "brands": ["Sony", "Bose", "Apple"],
    "features": ["noise_cancelling", "wireless", "over_ear"],
    "sort_by": "match_score",
    "limit": 20
  }
}

Response 200:
{
  "query": "wireless headphones noise cancelling",
  "total_results": 156,
  "products": [
    {
      "id": "SONY-WH1000XM5",
      "name": "Sony WH-1000XM5",
      "category": "Electronics > Audio > Headphones",
      "price": 349.99,
      "original_price": 399.99,
      "currency": "USD",
      "rating": 4.7,
      "reviews": 1523,
      "match_score": 0.98,
      "images": [
        "https://cdn.shopmindai.com/products/sony-wh1000xm5-1.jpg"
      ],
      "key_features": [
        "Industry-leading noise cancellation",
        "30-hour battery life",
        "Multipoint connection"
      ],
      "availability": {
        "in_stock": true,
        "stores": 5
      }
    }
  ],
  "facets": {
    "brands": {
      "Sony": 23,
      "Bose": 18,
      "Apple": 3
    },
    "price_ranges": {
      "100-200": 45,
      "200-300": 67,
      "300-400": 44
    }
  }
}
```

### 📊 Analytics & Insights

#### Get Shopping Insights
```http
GET /api/v1/users/{user_id}/insights
Authorization: Bearer <token>

Response 200:
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "period": "last_30_days",
  "insights": {
    "spending_pattern": {
      "total_spent": 2345.67,
      "average_per_purchase": 234.56,
      "categories": {
        "electronics": 1200.00,
        "fashion": 645.67,
        "home": 500.00
      }
    },
    "savings": {
      "total_saved": 456.78,
      "best_deal": {
        "product": "MacBook Air M2",
        "saved": 150.00,
        "discount": "12%"
      }
    },
    "preferences": {
      "favorite_brands": ["Apple", "Nike", "Samsung"],
      "preferred_retailers": ["Amazon", "Best Buy"],
      "shopping_times": {
        "most_active_day": "Saturday",
        "most_active_hour": "20:00"
      }
    },
    "recommendations": {
      "based_on_history": [
        "You tend to buy electronics during sales",
        "Consider setting price alerts for your wishlist items"
      ]
    }
  }
}
```

### 🔔 Notifications & Alerts

#### Set Price Alert
```http
POST /api/v1/alerts
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "SONY-WH1000XM5",
  "target_price": 299.99,
  "notification_channels": ["email", "push"]
}

Response 201:
{
  "alert_id": "880e8400-e29b-41d4-a716-446655440004",
  "product_id": "SONY-WH1000XM5",
  "current_price": 349.99,
  "target_price": 299.99,
  "status": "active",
  "created_at": "2024-01-20T15:35:00Z"
}
```

## Error Responses

All endpoints follow consistent error response format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-20T15:36:00Z"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request parameters |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/api/v1/auth/*` | 5 requests | 1 minute |
| `/api/v1/search` | 100 requests | 1 minute |
| `/api/v1/chat/*` | 50 requests | 1 minute |
| All others | 1000 requests | 1 minute |

## Webhooks

Configure webhooks to receive real-time events:

```http
POST /api/v1/webhooks
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://your-server.com/webhook",
  "events": ["price_drop", "back_in_stock", "new_recommendation"],
  "secret": "your_webhook_secret"
}
```

### Webhook Payload Example
```json
{
  "event": "price_drop",
  "data": {
    "product_id": "SONY-WH1000XM5",
    "old_price": 349.99,
    "new_price": 299.99,
    "discount": "14%",
    "retailer": "Amazon"
  },
  "timestamp": "2024-01-20T16:00:00Z",
  "signature": "sha256=..."
}
``` 

## Base URLs

- **Development:** `http://localhost:8080`
- **Production:** `https://api.shopmindai.com`
- **WebSocket:** `ws://localhost:8082/ws` (Chat Service)

## Authentication

All API requests (except `/auth/*` endpoints) require JWT authentication:

```http
Authorization: Bearer <JWT_TOKEN>
```

## API Endpoints

### 🔐 Authentication Service

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}

Response 201:
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2024-01-20T15:30:00Z"
}

Response 400:
{
  "error": "Email already exists"
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response 200:
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe"
  }
}

Response 401:
{
  "error": "Invalid credentials"
}
```

### 👤 User Service

#### Create User Profile
```http
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "preferences": {
    "categories": ["electronics", "fashion", "home"],
    "budget": {
      "min": 100,
      "max": 5000,
      "currency": "USD"
    },
    "brands": ["Apple", "Samsung", "Nike"],
    "notifications": {
      "email": true,
      "push": true,
      "sms": false
    }
  }
}

Response 201:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "preferences": {...},
  "created_at": "2024-01-20T15:30:00Z",
  "updated_at": "2024-01-20T15:30:00Z"
}
```

#### Get User Profile
```http
GET /api/v1/users/{user_id}
Authorization: Bearer <token>

Response 200:
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "preferences": {
    "categories": ["electronics", "fashion"],
    "budget": {
      "min": 100,
      "max": 5000,
      "currency": "USD"
    },
    "brands": ["Apple", "Samsung"],
    "notifications": {
      "email": true,
      "push": true,
      "sms": false
    }
  },
  "stats": {
    "total_searches": 145,
    "products_viewed": 523,
    "products_purchased": 12,
    "money_saved": 234.50
  },
  "created_at": "2024-01-20T15:30:00Z",
  "updated_at": "2024-01-25T10:15:00Z"
}
```

### 💬 Chat Service

#### Create Chat Session
```http
POST /api/v1/chat/sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Shopping for a new laptop",
  "context": {
    "budget": 1500,
    "purpose": "gaming and work"
  }
}

Response 201:
{
  "session_id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Shopping for a new laptop",
  "created_at": "2024-01-20T15:30:00Z"
}
```

#### Send Message
```http
POST /api/v1/chat/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "session_id": "660e8400-e29b-41d4-a716-446655440001",
  "content": "Find me a gaming laptop under $1500 with RTX 4060",
  "attachments": []
}

Response 200:
{
  "message_id": "770e8400-e29b-41d4-a716-446655440002",
  "session_id": "660e8400-e29b-41d4-a716-446655440001",
  "role": "user",
  "content": "Find me a gaming laptop under $1500 with RTX 4060",
  "timestamp": "2024-01-20T15:31:00Z",
  "ai_response": {
    "message_id": "770e8400-e29b-41d4-a716-446655440003",
    "content": "I found 3 excellent gaming laptops that match your criteria...",
    "recommendations": [
      {
        "product_id": "ASUS-ROG-G15",
        "name": "ASUS ROG Strix G15",
        "price": 1299.99,
        "original_price": 1499.99,
        "discount": "13%",
        "specs": {
          "gpu": "RTX 4060",
          "cpu": "AMD Ryzen 7 7735HS",
          "ram": "16GB DDR5",
          "storage": "512GB NVMe SSD"
        },
        "rating": 4.5,
        "reviews": 234,
        "match_score": 0.95,
        "retailers": [
          {
            "name": "Amazon",
            "price": 1299.99,
            "in_stock": true,
            "delivery": "2-day shipping",
            "url": "https://amazon.com/..."
          },
          {
            "name": "Best Buy",
            "price": 1349.99,
            "in_stock": true,
            "delivery": "In-store pickup available",
            "url": "https://bestbuy.com/..."
          }
        ]
      }
    ],
    "analysis": {
      "pros": [
        "Excellent performance for gaming and work",
        "RTX 4060 can handle modern games at 1080p/1440p",
        "Good cooling system"
      ],
      "cons": [
        "Battery life is average (4-5 hours)",
        "Can get loud under heavy load"
      ],
      "best_deal": {
        "retailer": "Amazon",
        "savings": 200.00,
        "reason": "Lowest price + fast shipping"
      }
    }
  }
}
```

#### WebSocket Connection (Real-time Chat)
```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8082/ws?session_id=660e8400-e29b-41d4-a716-446655440001');

// Send message
ws.send(JSON.stringify({
  type: 'message',
  content: 'Show me more options with better battery life'
}));

// Receive streaming response
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch(data.type) {
    case 'stream':
      // Partial response (real-time typing effect)
      console.log('AI typing:', data.chunk);
      break;
      
    case 'recommendation':
      // Product recommendation
      console.log('New product:', data.product);
      break;
      
    case 'complete':
      // Full response ready
      console.log('AI response complete:', data.content);
      break;
      
    case 'price_alert':
      // Real-time price drop notification
      console.log('Price dropped!', data.product, data.new_price);
      break;
  }
};
```

### 🔍 Search & Recommendations

#### Search Products
```http
POST /api/v1/search
Authorization: Bearer <token>
Content-Type: application/json

{
  "query": "wireless headphones noise cancelling",
  "filters": {
    "price_range": {
      "min": 100,
      "max": 400
    },
    "brands": ["Sony", "Bose", "Apple"],
    "features": ["noise_cancelling", "wireless", "over_ear"],
    "sort_by": "match_score",
    "limit": 20
  }
}

Response 200:
{
  "query": "wireless headphones noise cancelling",
  "total_results": 156,
  "products": [
    {
      "id": "SONY-WH1000XM5",
      "name": "Sony WH-1000XM5",
      "category": "Electronics > Audio > Headphones",
      "price": 349.99,
      "original_price": 399.99,
      "currency": "USD",
      "rating": 4.7,
      "reviews": 1523,
      "match_score": 0.98,
      "images": [
        "https://cdn.shopmindai.com/products/sony-wh1000xm5-1.jpg"
      ],
      "key_features": [
        "Industry-leading noise cancellation",
        "30-hour battery life",
        "Multipoint connection"
      ],
      "availability": {
        "in_stock": true,
        "stores": 5
      }
    }
  ],
  "facets": {
    "brands": {
      "Sony": 23,
      "Bose": 18,
      "Apple": 3
    },
    "price_ranges": {
      "100-200": 45,
      "200-300": 67,
      "300-400": 44
    }
  }
}
```

### 📊 Analytics & Insights

#### Get Shopping Insights
```http
GET /api/v1/users/{user_id}/insights
Authorization: Bearer <token>

Response 200:
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "period": "last_30_days",
  "insights": {
    "spending_pattern": {
      "total_spent": 2345.67,
      "average_per_purchase": 234.56,
      "categories": {
        "electronics": 1200.00,
        "fashion": 645.67,
        "home": 500.00
      }
    },
    "savings": {
      "total_saved": 456.78,
      "best_deal": {
        "product": "MacBook Air M2",
        "saved": 150.00,
        "discount": "12%"
      }
    },
    "preferences": {
      "favorite_brands": ["Apple", "Nike", "Samsung"],
      "preferred_retailers": ["Amazon", "Best Buy"],
      "shopping_times": {
        "most_active_day": "Saturday",
        "most_active_hour": "20:00"
      }
    },
    "recommendations": {
      "based_on_history": [
        "You tend to buy electronics during sales",
        "Consider setting price alerts for your wishlist items"
      ]
    }
  }
}
```

### 🔔 Notifications & Alerts

#### Set Price Alert
```http
POST /api/v1/alerts
Authorization: Bearer <token>
Content-Type: application/json

{
  "product_id": "SONY-WH1000XM5",
  "target_price": 299.99,
  "notification_channels": ["email", "push"]
}

Response 201:
{
  "alert_id": "880e8400-e29b-41d4-a716-446655440004",
  "product_id": "SONY-WH1000XM5",
  "current_price": 349.99,
  "target_price": 299.99,
  "status": "active",
  "created_at": "2024-01-20T15:35:00Z"
}
```

## Error Responses

All endpoints follow consistent error response format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-20T15:36:00Z"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Invalid request parameters |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/api/v1/auth/*` | 5 requests | 1 minute |
| `/api/v1/search` | 100 requests | 1 minute |
| `/api/v1/chat/*` | 50 requests | 1 minute |
| All others | 1000 requests | 1 minute |

## Webhooks

Configure webhooks to receive real-time events:

```http
POST /api/v1/webhooks
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://your-server.com/webhook",
  "events": ["price_drop", "back_in_stock", "new_recommendation"],
  "secret": "your_webhook_secret"
}
```

### Webhook Payload Example
```json
{
  "event": "price_drop",
  "data": {
    "product_id": "SONY-WH1000XM5",
    "old_price": 349.99,
    "new_price": 299.99,
    "discount": "14%",
    "retailer": "Amazon"
  },
  "timestamp": "2024-01-20T16:00:00Z",
  "signature": "sha256=..."
}
``` 