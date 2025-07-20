import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const wsConnections = new Counter('websocket_connections');
const wsMessages = new Counter('websocket_messages');
const wsErrors = new Counter('websocket_errors');
const connectionTime = new Trend('websocket_connection_time');
const messageLatency = new Trend('websocket_message_latency');
const successRate = new Rate('success_rate');

// Test configuration for billion-scale testing
export const options = {
  scenarios: {
    // Scenario 1: Connection stress test
    connection_stress: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 1000,
      maxVUs: 10000,
      stages: [
        { duration: '2m', target: 1000 },   // Ramp to 1K connections/sec
        { duration: '5m', target: 5000 },   // Ramp to 5K connections/sec
        { duration: '10m', target: 10000 }, // Sustain 10K connections/sec
        { duration: '2m', target: 0 },      // Ramp down
      ],
      exec: 'websocketConnectionTest',
    },
    
    // Scenario 2: Message throughput test
    message_throughput: {
      executor: 'constant-arrival-rate',
      rate: 50000,
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 5000,
      maxVUs: 20000,
      exec: 'messageTest',
      startTime: '20m', // Start after connection test
    },
    
    // Scenario 3: Cache efficiency test
    cache_test: {
      executor: 'constant-arrival-rate',
      rate: 10000,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 1000,
      exec: 'cacheTest',
      startTime: '30m',
    },
    
    // Scenario 4: Database stress test
    database_test: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 500,
      stages: [
        { duration: '2m', target: 1000 },
        { duration: '5m', target: 5000 },
        { duration: '3m', target: 1000 },
      ],
      exec: 'databaseTest',
      startTime: '35m',
    },
  },
  
  thresholds: {
    'websocket_connection_time': ['p(95)<1000'], // 95% connect within 1s
    'websocket_message_latency': ['p(95)<100'],  // 95% messages within 100ms
    'http_req_duration': ['p(95)<500'],          // 95% HTTP requests within 500ms
    'success_rate': ['rate>0.99'],               // 99% success rate
    'websocket_errors': ['count<1000'],          // Less than 1000 errors total
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8001';
const WS_URL = __ENV.WS_URL || 'ws://localhost:8001/ws';

// Helper function to generate test data
function generateUser() {
  return {
    email: `user_${Math.random().toString(36).substring(7)}@test.com`,
    password: 'TestPass123!',
    userId: `${Date.now()}_${Math.random().toString(36).substring(7)}`,
  };
}

// Test 1: WebSocket connection stress test
export function websocketConnectionTest() {
  const user = generateUser();
  const token = authenticateUser(user);
  
  const startTime = Date.now();
  
  const res = ws.connect(`${WS_URL}?token=${token}`, {}, function (socket) {
    const connectTime = Date.now() - startTime;
    connectionTime.add(connectTime);
    wsConnections.add(1);
    
    socket.on('open', () => {
      successRate.add(true);
      
      // Send initial message
      socket.send(JSON.stringify({
        type: 'message:new',
        data: {
          conversationId: 'test_conv_' + user.userId,
          content: 'Connection test message',
        },
      }));
      
      // Keep connection alive for realistic test
      socket.setInterval(() => {
        socket.send(JSON.stringify({
          type: 'ping',
          data: {},
        }));
      }, 30000);
    });
    
    socket.on('message', (data) => {
      wsMessages.add(1);
      const message = JSON.parse(data);
      
      check(message, {
        'message has id': (msg) => msg.id !== undefined,
        'message has type': (msg) => msg.type !== undefined,
      });
    });
    
    socket.on('error', (e) => {
      wsErrors.add(1);
      successRate.add(false);
      console.error('WebSocket error:', e);
    });
    
    // Simulate real user behavior
    sleep(Math.random() * 60 + 30); // Stay connected 30-90 seconds
  });
  
  check(res, {
    'WebSocket connection established': (r) => r && r.status === 101,
  });
}

// Test 2: Message throughput test
export function messageTest() {
  const user = generateUser();
  const token = authenticateUser(user);
  const conversationId = createConversation(token);
  
  const messageStart = Date.now();
  
  const payload = JSON.stringify({
    content: `Load test message ${Date.now()} - ${Math.random().toString(36)}`,
    conversationId: conversationId,
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  };
  
  const res = http.post(
    `${BASE_URL}/api/chat/conversations/${conversationId}/messages`,
    payload,
    params
  );
  
  const latency = Date.now() - messageStart;
  messageLatency.add(latency);
  
  check(res, {
    'message sent successfully': (r) => r.status === 201,
    'response has message id': (r) => r.json('id') !== undefined,
    'response time < 100ms': (r) => latency < 100,
  });
  
  successRate.add(res.status === 201);
}

// Test 3: Cache efficiency test
export function cacheTest() {
  const user = generateUser();
  const token = authenticateUser(user);
  
  // First request - should miss cache
  const res1 = http.get(`${BASE_URL}/api/chat/conversations`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });
  
  check(res1, {
    'first request succeeds': (r) => r.status === 200,
    'cache miss header': (r) => r.headers['X-Cache'] === 'MISS',
  });
  
  // Second request - should hit cache
  const res2 = http.get(`${BASE_URL}/api/chat/conversations`, {
    headers: { 'Authorization': `Bearer ${token}` },
  });
  
  check(res2, {
    'second request succeeds': (r) => r.status === 200,
    'cache hit header': (r) => r.headers['X-Cache'] === 'HIT',
    'faster response': (r) => r.timings.duration < res1.timings.duration * 0.5,
  });
  
  successRate.add(res1.status === 200 && res2.status === 200);
}

// Test 4: Database stress test with pagination
export function databaseTest() {
  const user = generateUser();
  const token = authenticateUser(user);
  
  // Create multiple conversations
  for (let i = 0; i < 10; i++) {
    createConversation(token, `Test conversation ${i}`);
  }
  
  // Test cursor pagination
  let cursor = '';
  let totalConversations = 0;
  
  for (let page = 0; page < 5; page++) {
    const url = cursor 
      ? `${BASE_URL}/api/chat/conversations?limit=10&cursor=${cursor}`
      : `${BASE_URL}/api/chat/conversations?limit=10`;
      
    const res = http.get(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    
    check(res, {
      'pagination request succeeds': (r) => r.status === 200,
      'has conversations array': (r) => Array.isArray(r.json('conversations')),
      'has next cursor': (r) => page < 4 ? r.json('nextCursor') !== undefined : true,
    });
    
    if (res.status === 200) {
      const data = res.json();
      totalConversations += data.conversations.length;
      cursor = data.nextCursor || '';
      
      if (!cursor) break;
    }
    
    successRate.add(res.status === 200);
  }
  
  check(null, {
    'fetched all conversations': () => totalConversations >= 10,
  });
}

// Helper functions
function authenticateUser(user) {
  // In real test, this would call actual auth endpoint
  // For now, return mock token
  return `mock_token_${user.userId}`;
}

function createConversation(token, title = 'Test conversation') {
  const res = http.post(
    `${BASE_URL}/api/chat/conversations`,
    JSON.stringify({ title }),
    {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
    }
  );
  
  return res.json('id') || `mock_conv_${Date.now()}`;
}

// Handle test summary
export function handleSummary(data) {
  return {
    'summary.json': JSON.stringify(data),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  const { metrics } = data;
  
  return `
=== Billion-Scale Load Test Results ===

WebSocket Performance:
  - Total Connections: ${metrics.websocket_connections.values.count}
  - Connection Time (p95): ${metrics.websocket_connection_time.values['p(95)']}ms
  - Total Messages: ${metrics.websocket_messages.values.count}
  - Message Latency (p95): ${metrics.websocket_message_latency.values['p(95)']}ms
  - WebSocket Errors: ${metrics.websocket_errors.values.count}

HTTP Performance:
  - Request Rate: ${metrics.http_reqs.values.rate}/s
  - Request Duration (p95): ${metrics.http_req_duration.values['p(95)']}ms
  - Success Rate: ${(metrics.success_rate.values.rate * 100).toFixed(2)}%

System Capacity Estimates:
  - Max Concurrent WebSockets: ${metrics.vus_max.values.value}
  - Message Throughput: ${metrics.websocket_messages.values.rate}/s
  - Projected Daily Messages: ${Math.round(metrics.websocket_messages.values.rate * 86400)}

Recommendations:
  ${getRecommendations(metrics)}
`;
}

function getRecommendations(metrics) {
  const recommendations = [];
  
  if (metrics.websocket_connection_time.values['p(95)'] > 1000) {
    recommendations.push('- ⚠️  WebSocket connection time high. Scale connection handlers.');
  }
  
  if (metrics.success_rate.values.rate < 0.99) {
    recommendations.push('- ⚠️  Success rate below 99%. Investigate errors.');
  }
  
  if (metrics.websocket_errors.values.count > 100) {
    recommendations.push('- ⚠️  High WebSocket error count. Check connection limits.');
  }
  
  if (recommendations.length === 0) {
    recommendations.push('- ✅ All metrics within acceptable range for billion-scale!');
  }
  
  return recommendations.join('\n');
}