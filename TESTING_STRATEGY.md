# 🧪 ShopGPT Microservices Testing Strategy

## 📋 Overview

This document outlines our comprehensive testing strategy for the ShopGPT microservices architecture, based on industry best practices and the Test Pyramid approach.

## 🔺 Testing Pyramid

```
         /\
        /E2E\        (5%)  - End-to-End Tests
       /------\
      /Contract\     (10%) - Contract Tests  
     /----------\
    /Integration \   (20%) - Integration Tests
   /--------------\
  / Component Tests\ (25%) - Component Tests
 /------------------\
/    Unit Tests      \ (40%) - Unit Tests
```

## 📊 Test Types & Strategies

### 1. **Unit Testing** (40% of tests)

**Purpose**: Test individual functions and methods in isolation

**Tools**: 
- Go: `testing` package, `testify`, `gomock`
- Frontend: Jest, React Testing Library

**Example**:
```go
// services/chat-service/internal/handlers/chat_handler_test.go
func TestValidateMessage(t *testing.T) {
    tests := []struct {
        name    string
        message string
        wantErr bool
    }{
        {
            name:    "valid message",
            message: "Hello ShopGPT",
            wantErr: false,
        },
        {
            name:    "empty message",
            message: "",
            wantErr: true,
        },
        {
            name:    "message too long",
            message: strings.Repeat("a", 5001),
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateMessage(tt.message)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 2. **Component Testing** (25% of tests)

**Purpose**: Test a microservice in isolation with mocked dependencies

**Tools**: 
- `httptest` for HTTP testing
- `testcontainers-go` for dependencies
- `gomock` for mocking

**Example**:
```go
// services/user-service/component_test.go
func TestUserServiceComponent(t *testing.T) {
    // Start test containers
    ctx := context.Background()
    postgres, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
            },
        },
    })
    defer postgres.Terminate(ctx)
    
    // Setup service with test DB
    service := NewUserService(postgres.ConnectionString())
    
    // Test user creation
    user, err := service.CreateUser(ctx, CreateUserRequest{
        Email:    "test@shopgpt.com",
        Username: "testuser",
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### 3. **Integration Testing** (20% of tests)

**Purpose**: Test interaction between multiple microservices

**Tools**:
- Docker Compose for service orchestration
- `net/http` for API testing
- Postman/Newman for automated API tests

**Example Docker Compose for Integration Tests**:
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  user-service:
    build: ./services/user-service
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
      
  chat-service:
    build: ./services/chat-service
    environment:
      - USER_SERVICE_URL=http://user-service:8080
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - kafka
      - user-service
      
  integration-tests:
    build: ./tests/integration
    command: go test -v ./...
    depends_on:
      - user-service
      - chat-service
```

### 4. **Contract Testing** (10% of tests)

**Purpose**: Ensure services maintain compatible APIs

**Tools**: 
- Pact for consumer-driven contracts
- OpenAPI/Swagger for API documentation

**Example Pact Test**:
```go
// Consumer test (chat-service)
func TestChatServiceUserContract(t *testing.T) {
    pact := &dsl.Pact{
        Consumer: "ChatService",
        Provider: "UserService",
    }
    
    pact.AddInteraction().
        Given("User with ID 123 exists").
        UponReceiving("A request for user details").
        WithRequest(dsl.Request{
            Method: "GET",
            Path:   dsl.String("/api/v1/users/123"),
        }).
        WillRespondWith(dsl.Response{
            Status: 200,
            Body: dsl.Like(User{
                ID:       "123",
                Username: "testuser",
                Email:    "test@example.com",
            }),
        })
        
    // Test the interaction
    err := pact.Verify(func() error {
        client := NewUserServiceClient(pact.Server.URL)
        user, err := client.GetUser("123")
        assert.NoError(t, err)
        assert.Equal(t, "123", user.ID)
        return nil
    })
    
    assert.NoError(t, err)
}
```

### 5. **End-to-End Testing** (5% of tests)

**Purpose**: Test complete user journeys across all services

**Tools**:
- Cypress for frontend E2E
- k6 for API load testing
- Playwright for cross-browser testing

**Example E2E Test**:
```javascript
// tests/e2e/shopping-journey.spec.js
describe('Complete Shopping Journey', () => {
  it('should search, chat, and get product recommendations', () => {
    cy.visit('/')
    
    // Login
    cy.get('[data-testid="login-button"]').click()
    cy.get('[name="email"]').type('test@shopgpt.com')
    cy.get('[name="password"]').type('password123')
    cy.get('[type="submit"]').click()
    
    // Search for product
    cy.get('[data-testid="search-input"]').type('gaming laptop')
    cy.get('[data-testid="store-selector"]').select('All stores')
    cy.get('[data-testid="search-button"]').click()
    
    // Verify results
    cy.get('[data-testid="product-card"]').should('have.length.greaterThan', 0)
    cy.get('[data-testid="chat-message"]').should('contain', 'found')
    
    // Interact with results
    cy.get('[data-testid="product-card"]').first().click()
    cy.get('[data-testid="add-to-compare"]').click()
  })
})
```

## 🔧 Testing Tools Stack

### Backend Testing
- **Unit**: Go testing, testify, gomock
- **Integration**: testcontainers-go, httptest
- **Contract**: Pact
- **Performance**: k6, vegeta
- **Security**: gosec, trivy

### Frontend Testing
- **Unit**: Jest, React Testing Library
- **Integration**: MSW (Mock Service Worker)
- **E2E**: Cypress, Playwright
- **Visual**: Chromatic, Percy
- **Accessibility**: jest-axe, Pa11y

### Infrastructure Testing
- **Container**: Container Structure Test
- **Kubernetes**: kubeval, conftest
- **Terraform**: terratest

## 🚀 CI/CD Integration

### GitHub Actions Workflow
```yaml
name: Test Pipeline
on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [user-service, chat-service, auth-service]
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run Unit Tests
        run: |
          cd services/${{ matrix.service }}
          go test -v -cover -race ./...
          
  component-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - uses: actions/checkout@v3
      - name: Run Component Tests
        run: |
          docker-compose -f docker-compose.test.yml up --abort-on-container-exit
          
  contract-tests:
    runs-on: ubuntu-latest
    needs: component-tests
    steps:
      - name: Run Contract Tests
        run: |
          make contract-tests
          
  e2e-tests:
    runs-on: ubuntu-latest
    needs: contract-tests
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy to staging
        run: kubectl apply -f k8s/staging/
      - name: Run E2E Tests
        run: npm run test:e2e
```

## 📊 Test Coverage Goals

| Test Type | Coverage Goal | Current | Status |
|-----------|--------------|---------|---------|
| Unit | 80% | 0% | 🔴 |
| Component | 70% | 0% | 🔴 |
| Integration | 60% | 0% | 🔴 |
| Contract | 100% | 0% | 🔴 |
| E2E | Critical paths | 0% | 🔴 |

## 🏃 Performance Testing

### Load Testing Strategy
```yaml
# k6/scenarios/chat-load.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<1500'], // 99% of requests under 1.5s
    http_req_failed: ['rate<0.1'],     // Error rate under 10%
  },
};

export default function () {
  const payload = JSON.stringify({
    message: 'Find me a laptop under $1000',
    store: 'all',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${__ENV.API_TOKEN}`,
    },
  };

  const res = http.post('http://api.shopgpt.com/v1/chat', payload, params);
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1500ms': (r) => r.timings.duration < 1500,
    'has products': (r) => JSON.parse(r.body).products.length > 0,
  });
  
  sleep(1);
}
```

## 🔒 Security Testing

### OWASP Top 10 Coverage
- [ ] Injection attacks (SQLi, NoSQLi)
- [ ] Broken Authentication
- [ ] Sensitive Data Exposure
- [ ] XML External Entities (XXE)
- [ ] Broken Access Control
- [ ] Security Misconfiguration
- [ ] Cross-Site Scripting (XSS)
- [ ] Insecure Deserialization
- [ ] Using Components with Known Vulnerabilities
- [ ] Insufficient Logging & Monitoring

### Security Test Example
```go
// services/auth-service/security_test.go
func TestSQLInjectionPrevention(t *testing.T) {
    maliciousInputs := []string{
        "'; DROP TABLE users; --",
        "1' OR '1'='1",
        "admin'--",
    }
    
    for _, input := range maliciousInputs {
        _, err := authService.Login(input, "password")
        assert.Error(t, err)
        assert.NotContains(t, err.Error(), "syntax error")
    }
}
```

## 🤖 Test Automation

### Automated Test Generation
```bash
# Generate unit tests for Go code
gotests -all -w services/chat-service/internal/handlers/

# Generate API tests from OpenAPI spec
openapi-generator generate -i api/openapi.yaml -g go -o tests/generated/

# Generate load tests from HAR files
har-to-k6 recording.har -o k6/scenarios/recorded.js
```

## 📈 Monitoring & Observability in Tests

### Test Metrics to Track
- Test execution time
- Flaky test rate
- Code coverage trends
- Test failure reasons
- Performance regression detection

### Grafana Dashboard for Tests
```json
{
  "dashboard": {
    "title": "Test Metrics",
    "panels": [
      {
        "title": "Test Success Rate",
        "targets": [
          {
            "expr": "rate(test_runs_total{status=\"passed\"}[5m]) / rate(test_runs_total[5m])"
          }
        ]
      },
      {
        "title": "Average Test Duration",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, test_duration_seconds)"
          }
        ]
      }
    ]
  }
}
```

## 🎯 Testing Best Practices

1. **Test Isolation**: Each test should be independent
2. **Deterministic**: Tests should produce same results every time
3. **Fast Feedback**: Prioritize fast tests in CI pipeline
4. **Clear Naming**: Test names should describe what they test
5. **Arrange-Act-Assert**: Follow AAA pattern
6. **Test Data Management**: Use factories and fixtures
7. **Parallel Execution**: Run tests in parallel when possible
8. **Continuous Testing**: Test on every commit

## 📅 Testing Roadmap

### Phase 1: Foundation (Weeks 1-2)
- [ ] Set up testing infrastructure
- [ ] Implement unit tests for critical paths
- [ ] Configure CI pipeline

### Phase 2: Integration (Weeks 3-4)
- [ ] Add component tests
- [ ] Implement contract testing
- [ ] Set up test containers

### Phase 3: E2E & Performance (Weeks 5-6)
- [ ] Implement E2E test suite
- [ ] Add performance tests
- [ ] Security testing integration

### Phase 4: Optimization (Ongoing)
- [ ] Improve test coverage
- [ ] Reduce test execution time
- [ ] Implement test result analytics