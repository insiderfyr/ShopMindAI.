import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 100 },   // Ramp up to 100 users
    { duration: '1m', target: 500 },    // Ramp up to 500 users
    { duration: '2m', target: 1000 },   // Ramp up to 1000 users
    { duration: '2m', target: 1000 },   // Stay at 1000 users
    { duration: '1m', target: 0 },      // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'], // 99% of requests must complete below 500ms
    errors: ['rate<0.01'],            // Error rate must be below 1%
  },
};

const BASE_URL = 'http://localhost:8080';

// Test data
const testUser = {
  email: `test${Math.random()}@example.com`,
  password: 'Test123!',
  name: 'Load Test User'
};

export function setup() {
  // Register a test user
  const res = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (res.status !== 201) {
    throw new Error('Setup failed: Could not register test user');
  }
  
  // Login to get token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: testUser.email,
    password: testUser.password
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const token = JSON.parse(loginRes.body).access_token;
  return { token, userId: JSON.parse(res.body).user_id };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`
  };

  // Scenario 1: Get user profile (70% of traffic)
  if (Math.random() < 0.7) {
    const res = http.get(`${BASE_URL}/api/v1/users/${data.userId}`, { headers });
    
    check(res, {
      'Get user profile - status is 200': (r) => r.status === 200,
      'Get user profile - has user data': (r) => JSON.parse(r.body).id === data.userId,
    });
    
    errorRate.add(res.status !== 200);
  }
  
  // Scenario 2: Send chat message (20% of traffic)
  else if (Math.random() < 0.9) {
    // Create chat session
    const sessionRes = http.post(`${BASE_URL}/api/v1/chat/sessions`, JSON.stringify({
      title: 'Load test session'
    }), { headers });
    
    if (sessionRes.status === 201) {
      const sessionId = JSON.parse(sessionRes.body).session_id;
      
      // Send message
      const messageRes = http.post(`${BASE_URL}/api/v1/chat/messages`, JSON.stringify({
        session_id: sessionId,
        content: 'Find me a laptop under $1000'
      }), { headers });
      
      check(messageRes, {
        'Send message - status is 200': (r) => r.status === 200,
        'Send message - has AI response': (r) => JSON.parse(r.body).ai_response !== undefined,
      });
      
      errorRate.add(messageRes.status !== 200);
    }
  }
  
  // Scenario 3: Search products (10% of traffic)
  else {
    const searchRes = http.post(`${BASE_URL}/api/v1/search`, JSON.stringify({
      query: 'wireless headphones',
      filters: {
        price_range: { min: 50, max: 300 },
        limit: 20
      }
    }), { headers });
    
    check(searchRes, {
      'Search products - status is 200': (r) => r.status === 200,
      'Search products - has results': (r) => JSON.parse(r.body).products.length > 0,
    });
    
    errorRate.add(searchRes.status !== 200);
  }
  
  sleep(1); // Think time between requests
}

export function teardown(data) {
  // Cleanup: delete test user
  http.del(`${BASE_URL}/api/v1/users/${data.userId}`, {
    headers: {
      'Authorization': `Bearer ${data.token}`
    }
  });
} 
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 100 },   // Ramp up to 100 users
    { duration: '1m', target: 500 },    // Ramp up to 500 users
    { duration: '2m', target: 1000 },   // Ramp up to 1000 users
    { duration: '2m', target: 1000 },   // Stay at 1000 users
    { duration: '1m', target: 0 },      // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'], // 99% of requests must complete below 500ms
    errors: ['rate<0.01'],            // Error rate must be below 1%
  },
};

const BASE_URL = 'http://localhost:8080';

// Test data
const testUser = {
  email: `test${Math.random()}@example.com`,
  password: 'Test123!',
  name: 'Load Test User'
};

export function setup() {
  // Register a test user
  const res = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify(testUser), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (res.status !== 201) {
    throw new Error('Setup failed: Could not register test user');
  }
  
  // Login to get token
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
    email: testUser.email,
    password: testUser.password
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  const token = JSON.parse(loginRes.body).access_token;
  return { token, userId: JSON.parse(res.body).user_id };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`
  };

  // Scenario 1: Get user profile (70% of traffic)
  if (Math.random() < 0.7) {
    const res = http.get(`${BASE_URL}/api/v1/users/${data.userId}`, { headers });
    
    check(res, {
      'Get user profile - status is 200': (r) => r.status === 200,
      'Get user profile - has user data': (r) => JSON.parse(r.body).id === data.userId,
    });
    
    errorRate.add(res.status !== 200);
  }
  
  // Scenario 2: Send chat message (20% of traffic)
  else if (Math.random() < 0.9) {
    // Create chat session
    const sessionRes = http.post(`${BASE_URL}/api/v1/chat/sessions`, JSON.stringify({
      title: 'Load test session'
    }), { headers });
    
    if (sessionRes.status === 201) {
      const sessionId = JSON.parse(sessionRes.body).session_id;
      
      // Send message
      const messageRes = http.post(`${BASE_URL}/api/v1/chat/messages`, JSON.stringify({
        session_id: sessionId,
        content: 'Find me a laptop under $1000'
      }), { headers });
      
      check(messageRes, {
        'Send message - status is 200': (r) => r.status === 200,
        'Send message - has AI response': (r) => JSON.parse(r.body).ai_response !== undefined,
      });
      
      errorRate.add(messageRes.status !== 200);
    }
  }
  
  // Scenario 3: Search products (10% of traffic)
  else {
    const searchRes = http.post(`${BASE_URL}/api/v1/search`, JSON.stringify({
      query: 'wireless headphones',
      filters: {
        price_range: { min: 50, max: 300 },
        limit: 20
      }
    }), { headers });
    
    check(searchRes, {
      'Search products - status is 200': (r) => r.status === 200,
      'Search products - has results': (r) => JSON.parse(r.body).products.length > 0,
    });
    
    errorRate.add(searchRes.status !== 200);
  }
  
  sleep(1); // Think time between requests
}

export function teardown(data) {
  // Cleanup: delete test user
  http.del(`${BASE_URL}/api/v1/users/${data.userId}`, {
    headers: {
      'Authorization': `Bearer ${data.token}`
    }
  });
} 