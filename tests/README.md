# LLM-Proxy Testing Suite

Comprehensive testing suite for the LLM-Proxy project including unit tests, integration tests, and load tests.

## 📁 Test Structure

```
tests/
├── integration/          # Integration tests (API endpoints)
│   └── api_test.go
├── load/                 # Load tests (k6)
│   ├── oauth-load-test.js
│   ├── api-endpoints-load-test.js
│   ├── stress-test.js
│   └── spike-test.js
└── README.md            # This file

internal/
├── application/
│   ├── oauth/
│   │   └── service_test.go         # OAuth service unit tests
│   └── caching/
│       └── service_test.go         # Caching service unit tests
└── infrastructure/
    └── providers/
        └── claude/
            └── mapper_test.go      # Claude mapper unit tests
```

---

## 🧪 Unit Tests

Unit tests test individual components in isolation using mocks.

### Run Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/application/oauth/

# Run with race detection
go test -race ./...

# Run with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...
```

### Unit Test Coverage

- **OAuth Service** (`internal/application/oauth/service_test.go`)
  - ✅ Token generation (client_credentials grant)
  - ✅ Invalid credentials handling
  - ✅ Inactive client handling
  - ✅ Token validation
  - ✅ Expired token detection
  - ✅ Token revocation
  - ✅ Token refresh flow
  - ✅ Benchmarks

- **Caching Service** (`internal/application/caching/service_test.go`)
  - ✅ Cache get/set operations
  - ✅ Cache misses and errors
  - ✅ Pattern-based invalidation
  - ✅ Cache key generation consistency
  - ✅ Statistics tracking
  - ✅ Benchmarks

- **Claude Mapper** (`internal/infrastructure/providers/claude/mapper_test.go`)
  - ✅ Request mapping (OpenAI → Claude)
  - ✅ Response mapping (Claude → OpenAI)
  - ✅ System message handling
  - ✅ Conversation history mapping
  - ✅ Cost calculation (all models)
  - ✅ Benchmarks

---

## 🔗 Integration Tests

Integration tests test complete API flows with real HTTP requests.

### Prerequisites

- Backend server running on http://localhost:8080
- Docker services (PostgreSQL, Redis) running
- Test OAuth client configured

### Setup

```bash
# Start Docker services
cd deployments/docker
docker compose up -d postgres redis

# Start backend server
cd ../..
make dev
# Or: go run cmd/server/main.go
```

### Run Integration Tests

```bash
# Run integration tests
go test -tags=integration ./tests/integration/

# Run with verbose output
go test -tags=integration -v ./tests/integration/

# Run specific test
go test -tags=integration -v -run TestHealth ./tests/integration/

# Skip short tests (run all including performance tests)
go test -tags=integration -v ./tests/integration/ -timeout 30m
```

### Environment Variables

Configure using environment variables:

```bash
export API_BASE_URL="http://localhost:8080"
export TEST_CLIENT_ID="test_client"
export TEST_CLIENT_SECRET="test_secret_123456"
export ADMIN_API_KEY="admin_dev_key_12345678901234567890123456789012"

# Then run tests
go test -tags=integration ./tests/integration/
```

### Integration Test Coverage

- ✅ Health endpoint
- ✅ OAuth token generation (client_credentials)
- ✅ OAuth invalid credentials
- ✅ OAuth token refresh
- ✅ List models endpoint
- ✅ Get single model endpoint
- ✅ Chat completions (authorization check)
- ✅ Admin: List clients
- ✅ Admin: Get single client
- ✅ Admin: Create client
- ✅ Admin: Cache stats
- ✅ Admin: Provider status
- ✅ Admin: Authorization check
- ✅ Metrics endpoint
- ✅ Performance: Multiple OAuth requests

---

## 🔥 Load Tests (k6)

Load tests simulate realistic traffic patterns to test system performance and limits.

### Prerequisites

Install k6:

```bash
# macOS (Homebrew)
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Or download binary from: https://k6.io/docs/get-started/installation/
```

### Run Load Tests

#### 1. OAuth Load Test
Tests OAuth token endpoint under increasing load:

```bash
cd tests/load

# Basic run
k6 run oauth-load-test.js

# With custom configuration
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e CLIENT_ID=test_client \
  -e CLIENT_SECRET=test_secret_123456 \
  oauth-load-test.js

# With output to InfluxDB (optional)
k6 run --out influxdb=http://localhost:8086/k6 oauth-load-test.js
```

**Test Profile:**
- Ramp up: 0 → 10 → 50 users
- Duration: ~4 minutes
- Target: 95% requests < 500ms

#### 2. API Endpoints Load Test
Realistic user behavior across multiple endpoints:

```bash
k6 run api-endpoints-load-test.js

# Or with all environment variables
k6 run \
  -e BASE_URL=http://localhost:8080 \
  -e CLIENT_ID=test_client \
  -e CLIENT_SECRET=test_secret_123456 \
  -e ADMIN_API_KEY=admin_dev_key_12345678901234567890123456789012 \
  api-endpoints-load-test.js
```

**Test Profile:**
- Realistic traffic: 5 → 10 → 30 RPS
- Duration: ~8 minutes
- Mix: Health (30%), Models (50%), Admin (20%)

#### 3. Stress Test
Gradually increases load to find breaking point:

```bash
k6 run stress-test.js
```

**Test Profile:**
- Progressive load: 20 → 50 → 100 → 150 → 200 users
- Duration: ~15 minutes
- Goal: Find system limits

#### 4. Spike Test
Tests system behavior under sudden traffic surges:

```bash
k6 run spike-test.js
```

**Test Profile:**
- Sudden spike: 10 → 200 users in 10 seconds
- Duration: ~5 minutes
- Goal: Test auto-scaling and recovery

### Load Test Results

View results in:
- **Console Output** - Real-time metrics during test
- **Grafana Dashboards** - http://localhost:3001
- **Prometheus** - http://localhost:9090

---

## 🚀 Running Tests in CI/CD

### GitLab CI

Tests are automatically run in GitLab CI pipeline:

```yaml
test:unit:
  stage: test
  script:
    - go test -v -race -coverprofile=coverage.txt ./...
  
test:integration:
  stage: test
  services:
    - postgres:14-alpine
    - redis:7-alpine
  script:
    - go test -tags=integration -v ./tests/integration/
```

See `.gitlab-ci.yml` for complete configuration.

---

## 📊 Test Metrics & Thresholds

### Unit Tests
- **Coverage Target:** >80%
- **Race Conditions:** Zero tolerance
- **Benchmark Regression:** <10% slower than baseline

### Integration Tests
- **Success Rate:** >99%
- **Response Time:** p95 < 1s, p99 < 2s
- **Timeout:** 30 minutes max

### Load Tests

#### OAuth Load Test
- http_req_duration: p95 < 500ms
- http_req_failed: < 1%
- errors: < 5%

#### API Endpoints Load Test
- http_req_duration: p95 < 1000ms, p99 < 2000ms
- http_req_failed: < 2%
- errors: < 5%

#### Stress Test
- http_req_duration: p95 < 2000ms, p99 < 5000ms
- http_req_failed: < 10%
- errors: < 15%

#### Spike Test
- http_req_duration: p95 < 3000ms
- http_req_failed: < 15%
- errors: < 20%

---

## 🐛 Troubleshooting

### Unit Tests Fail

**Issue:** Import errors or package not found
```bash
# Solution: Update dependencies
go mod tidy
go mod download
```

**Issue:** Test fails with "connection refused"
```bash
# Solution: Mock is not properly configured
# Check that you're using mocks, not real services
```

### Integration Tests Fail

**Issue:** Connection refused to http://localhost:8080
```bash
# Solution: Start backend server
make dev

# Or check if it's running
curl http://localhost:8080/health
```

**Issue:** Database connection errors
```bash
# Solution: Start Docker services
cd deployments/docker
docker compose up -d postgres redis
```

**Issue:** OAuth tests fail with 401
```bash
# Solution: Check test client exists in database
# Run migrations: make migrate-up
# Seed test data if needed
```

### Load Tests Fail

**Issue:** k6: command not found
```bash
# Solution: Install k6
# See installation instructions above
```

**Issue:** High error rate during load tests
```bash
# Solution: System may be under-resourced
# 1. Check Docker resource limits
# 2. Increase database connection pool
# 3. Scale backend if needed
```

**Issue:** Requests timeout
```bash
# Solution: Increase timeout in k6 script
# Or reduce load (lower VUs/RPS)
```

---

## 📚 Best Practices

### Writing Tests

1. **Use Table-Driven Tests**
   ```go
   tests := []struct{
       name string
       input string
       want string
   }{
       {"case1", "input1", "output1"},
       {"case2", "input2", "output2"},
   }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           got := function(tt.input)
           assert.Equal(t, tt.want, got)
       })
   }
   ```

2. **Use Meaningful Test Names**
   - Good: `TestOAuth_InvalidCredentials`
   - Bad: `TestOAuth1`

3. **Test Both Success and Failure Cases**
   - Happy path
   - Edge cases
   - Error conditions

4. **Use Setup/Teardown Functions**
   ```go
   func setupTest() (*Service, func()) {
       service := NewService(...)
       cleanup := func() { service.Close() }
       return service, cleanup
   }
   ```

5. **Avoid Test Interdependence**
   - Each test should be independent
   - Use `t.Parallel()` when possible

### Load Testing

1. **Start Small, Scale Up**
   - Begin with 1-10 VUs
   - Gradually increase load
   - Monitor system metrics

2. **Use Realistic Scenarios**
   - Mix different endpoints
   - Randomize request data
   - Include think time

3. **Monitor During Tests**
   - Watch Grafana dashboards
   - Check error logs
   - Monitor resource usage

4. **Document Results**
   - Record max throughput
   - Note bottlenecks
   - Save graphs/metrics

---

## 📖 Additional Resources

- **Go Testing:** https://go.dev/doc/tutorial/add-a-test
- **Testify:** https://github.com/stretchr/testify
- **k6 Documentation:** https://k6.io/docs/
- **Integration Testing:** https://martinfowler.com/bliki/IntegrationTest.html
- **Load Testing Best Practices:** https://k6.io/docs/test-types/

---

**Last Updated:** January 29, 2026
**Test Coverage:** Unit, Integration, Load
**Status:** ✅ Complete
