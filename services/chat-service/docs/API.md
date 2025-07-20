# Chat Service API Documentation

## Overview

The Chat Service provides real-time conversational AI capabilities for ShopGPT, enabling users to search for products, get recommendations, and interact with an AI shopping assistant. It uses WebSocket for real-time bidirectional communication and integrates with multiple store APIs.

## Architecture

- **Protocol**: WebSocket (with HTTP fallback)
- **Authentication**: JWT Bearer tokens
- **Message Format**: JSON
- **Real-time Features**: Typing indicators, streaming responses
- **AI Integration**: OpenAI GPT-4 / Claude 3

## WebSocket Connection

### Establishing Connection

**Endpoint**: `wss://api.shopgpt.com/v1/chat/ws`

**Authentication**: Include JWT token in connection headers or query parameter

```javascript
// JavaScript Example
const ws = new WebSocket('wss://api.shopgpt.com/v1/chat/ws', {
  headers: {
    'Authorization': 'Bearer <jwt_token>'
  }
});

// Alternative with query parameter
const ws = new WebSocket('wss://api.shopgpt.com/v1/chat/ws?token=<jwt_token>');
```

**Connection Response**: Upon successful connection, the server sends a welcome message:

```json
{
  "id": "msg-123-uuid",
  "type": "system",
  "content": "Welcome to ShopGPT! I'm here to help you find the best products across multiple stores. What are you looking for today?",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Connection Lifecycle

1. **Handshake**: Client connects with authentication
2. **Welcome**: Server sends welcome message
3. **Communication**: Bidirectional message exchange
4. **Heartbeat**: Ping/pong every 30 seconds
5. **Disconnection**: Clean closure or timeout

### Error Handling

Connection errors return appropriate HTTP status codes before upgrade:
- `401 Unauthorized` - Invalid or missing authentication
- `403 Forbidden` - User not allowed to access chat
- `429 Too Many Requests` - Rate limit exceeded
- `503 Service Unavailable` - Service temporarily down

## Message Protocol

### Message Structure

All messages follow this structure:

```typescript
interface Message {
  id: string;           // Unique message identifier
  type: MessageType;    // Type of message
  userId?: string;      // User ID (set by server)
  content?: string;     // Message content
  store?: string;       // Store filter (optional)
  products?: Product[]; // Product results (optional)
  error?: string;       // Error message (optional)
  metadata?: object;    // Additional data (optional)
  timestamp: string;    // ISO 8601 timestamp
}
```

### Message Types

#### 1. Chat Message (User → Server)

Send a conversational message to the AI assistant.

```json
{
  "type": "chat",
  "content": "I'm looking for a gaming laptop under $1500",
  "store": "all"  // Optional: "all", "amazon", "bestbuy", etc.
}
```

**Server Responses**:
1. Acknowledgment
2. Typing indicator
3. Assistant response with/without products

#### 2. Search Message (User → Server)

Explicitly search for products.

```json
{
  "type": "search",
  "content": "RTX 4070 graphics card",
  "store": "amazon",
  "filters": {
    "minPrice": 500,
    "maxPrice": 800,
    "inStock": true
  }
}
```

**Server Responses**:
1. Acknowledgment
2. Searching indicator
3. Search results with products

#### 3. Typing Indicator (Bidirectional)

Indicate typing status.

```json
{
  "type": "typing",
  "content": "user"  // or "assistant"
}
```

#### 4. System Messages (Server → User)

System notifications and updates.

```json
{
  "type": "system",
  "content": "Connection established",
  "metadata": {
    "serverVersion": "1.2.0",
    "features": ["chat", "search", "recommendations"]
  }
}
```

#### 5. Assistant Response (Server → User)

AI assistant's response to user queries.

```json
{
  "id": "msg-456-uuid",
  "type": "assistant",
  "content": "I found 5 excellent gaming laptops under $1500. Here are my top recommendations:\n\n1. **ASUS ROG Strix G15** - This laptop offers great performance...",
  "products": [
    {
      "id": "prod-789",
      "name": "ASUS ROG Strix G15 Gaming Laptop",
      "price": 1299.99,
      "store": "amazon",
      "url": "https://amazon.com/dp/B09XXX",
      "image": "https://cdn.shopgpt.com/products/asus-rog-g15.jpg",
      "description": "15.6\" 300Hz FHD, AMD Ryzen 9, RTX 3070, 16GB RAM, 1TB SSD",
      "inStock": true,
      "rating": 4.6,
      "reviews": 1247,
      "features": [
        "AMD Ryzen 9 6900HX",
        "NVIDIA RTX 3070 8GB",
        "16GB DDR5 RAM",
        "1TB NVMe SSD"
      ]
    }
  ],
  "timestamp": "2024-01-15T10:31:00Z"
}
```

#### 6. Search Results (Server → User)

Dedicated search results message.

```json
{
  "id": "msg-789-uuid",
  "type": "search_results",
  "content": "Found 12 products matching 'RTX 4070 graphics card'",
  "products": [...],
  "metadata": {
    "totalResults": 12,
    "searchTime": 1.23,
    "stores": ["amazon", "newegg", "bestbuy"]
  }
}
```

#### 7. Acknowledgment (Server → User)

Confirms message receipt.

```json
{
  "type": "ack",
  "content": "msg-123-uuid",  // ID of acknowledged message
  "timestamp": "2024-01-15T10:30:01Z"
}
```

#### 8. Error Message (Server → User)

Error notifications.

```json
{
  "type": "error",
  "error": "Failed to process request",
  "metadata": {
    "code": "PROCESSING_ERROR",
    "retryable": true,
    "details": "Temporary issue with product search API"
  }
}
```

### Product Object Structure

```typescript
interface Product {
  id: string;
  name: string;
  price: number;
  originalPrice?: number;    // If on sale
  store: string;
  url: string;
  image: string;
  description: string;
  inStock: boolean;
  rating?: number;          // 0-5
  reviews?: number;
  features?: string[];
  specifications?: object;
  shipping?: {
    free: boolean;
    prime?: boolean;
    estimatedDays: number;
  };
  seller?: {
    name: string;
    rating: number;
    verified: boolean;
  };
}
```

## HTTP REST API Endpoints

### 1. Get Chat History

Retrieve previous chat messages.

**Endpoint**: `GET /v1/chat/history`

**Query Parameters**:
- `limit` - Number of messages (default: 50, max: 200)
- `before` - Get messages before this timestamp
- `after` - Get messages after this timestamp

**Response**:
```json
{
  "messages": [
    {
      "id": "msg-123",
      "type": "chat",
      "content": "Show me gaming laptops",
      "timestamp": "2024-01-15T10:30:00Z",
      "userId": "user-123"
    },
    {
      "id": "msg-124",
      "type": "assistant",
      "content": "Here are the best gaming laptops...",
      "products": [...],
      "timestamp": "2024-01-15T10:30:05Z"
    }
  ],
  "hasMore": true,
  "total": 156
}
```

### 2. Delete Chat History

Clear chat history for privacy.

**Endpoint**: `DELETE /v1/chat/history`

**Response**: `204 No Content`

### 3. Export Chat

Export chat history in various formats.

**Endpoint**: `GET /v1/chat/export`

**Query Parameters**:
- `format` - Export format: `json`, `csv`, `pdf`
- `startDate` - Start date filter
- `endDate` - End date filter

**Response**: File download in requested format

### 4. Get Conversation Summary

Get AI-generated summary of a conversation.

**Endpoint**: `GET /v1/chat/conversations/{conversationId}/summary`

**Response**:
```json
{
  "conversationId": "conv-123",
  "summary": "User searched for gaming laptops under $1500. Recommended 5 options focusing on ASUS and MSI models with RTX 3070 graphics cards.",
  "keyProducts": ["prod-789", "prod-790"],
  "topics": ["gaming", "laptops", "budget"],
  "duration": 300,
  "messageCount": 12
}
```

## Advanced Features

### 1. Streaming Responses

For long AI responses, content is streamed progressively:

```json
{
  "type": "assistant",
  "content": "I found several great options",
  "streaming": true,
  "chunk": 1
}

{
  "type": "assistant",
  "content": " for gaming laptops. Let me show you",
  "streaming": true,
  "chunk": 2
}

{
  "type": "assistant",
  "content": " the best ones:\n\n1. ASUS ROG...",
  "streaming": false,
  "chunk": 3,
  "products": [...]
}
```

### 2. Multi-Store Search

Search across multiple stores simultaneously:

```json
{
  "type": "search",
  "content": "iPhone 15 Pro",
  "store": "all",
  "metadata": {
    "compareStores": true,
    "includeSellers": true
  }
}
```

### 3. Price Tracking

Set up price alerts:

```json
{
  "type": "price_alert",
  "productId": "prod-123",
  "targetPrice": 899.99,
  "duration": "30d"
}
```

### 4. Product Comparison

Compare multiple products:

```json
{
  "type": "compare",
  "productIds": ["prod-123", "prod-456", "prod-789"]
}
```

## Rate Limiting

WebSocket connections are rate limited:
- **Connection limit**: 5 concurrent connections per user
- **Message rate**: 30 messages per minute
- **Search rate**: 10 searches per minute

Rate limit information is sent in system messages:

```json
{
  "type": "system",
  "content": "Rate limit warning",
  "metadata": {
    "limit": 30,
    "remaining": 5,
    "resetAt": "2024-01-15T10:35:00Z"
  }
}
```

## Error Codes

| Code | Description | Retryable |
|------|-------------|-----------|
| `AUTH_FAILED` | Authentication failed | No |
| `RATE_LIMITED` | Rate limit exceeded | Yes (after reset) |
| `INVALID_MESSAGE` | Message format invalid | No |
| `PROCESSING_ERROR` | Server processing error | Yes |
| `SERVICE_UNAVAILABLE` | AI service unavailable | Yes |
| `STORE_API_ERROR` | Store API failure | Yes |
| `TIMEOUT` | Request timeout | Yes |

## Client Libraries

Official SDKs with WebSocket support:

### JavaScript/TypeScript

```javascript
import { ShopGPTChat } from '@shopgpt/chat-sdk';

const chat = new ShopGPTChat({
  token: 'your-jwt-token',
  onMessage: (message) => {
    console.log('Received:', message);
  }
});

await chat.connect();
await chat.sendMessage({
  type: 'chat',
  content: 'Find me a good coffee maker'
});
```

### Python

```python
from shopgpt import ChatClient

async with ChatClient(token='your-jwt-token') as chat:
    await chat.send_message(
        type='chat',
        content='Find me a good coffee maker'
    )
    
    async for message in chat.messages():
        print(f"Received: {message}")
```

## Best Practices

1. **Connection Management**
   - Implement exponential backoff for reconnection
   - Handle connection drops gracefully
   - Close connections when not in use

2. **Message Handling**
   - Always check message type before processing
   - Handle unknown message types gracefully
   - Implement timeout for responses

3. **Error Handling**
   - Display user-friendly error messages
   - Log errors for debugging
   - Implement retry logic for transient errors

4. **Performance**
   - Debounce typing indicators
   - Batch multiple searches when possible
   - Cache product data client-side

5. **Security**
   - Validate all incoming messages
   - Sanitize content before display
   - Refresh tokens before expiry

## Testing

WebSocket endpoint for testing:
- **URL**: `wss://api-test.shopgpt.com/v1/chat/ws`
- **Test Token**: Available in developer dashboard
- **Features**: Same as production with test data

## Support

- **Developer Portal**: https://developers.shopgpt.com/chat
- **API Status**: https://status.shopgpt.com
- **Support Email**: chat-api@shopgpt.com
- **Community**: https://community.shopgpt.com/chat-api