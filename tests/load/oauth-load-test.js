// =============================================================================
// K6 LOAD TEST - OAUTH TOKEN GENERATION
// =============================================================================
// Tests OAuth /oauth/token endpoint under load
// =============================================================================

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const tokenGenerationTime = new Trend('token_generation_time');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const CLIENT_ID = __ENV.CLIENT_ID || 'test_client';
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'test_secret_123456';

// Load test options
export const options = {
  scenarios: {
    // Scenario 1: Ramp-up load test
    ramp_up: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },  // Ramp up to 10 users over 30s
        { duration: '1m', target: 10 },   // Stay at 10 users for 1 minute
        { duration: '30s', target: 50 },  // Ramp up to 50 users over 30s
        { duration: '1m', target: 50 },   // Stay at 50 users for 1 minute
        { duration: '30s', target: 0 },   // Ramp down to 0 users
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<500'],  // 95% of requests should be below 500ms
    'http_req_failed': ['rate<0.01'],    // Error rate should be below 1%
    'errors': ['rate<0.05'],              // Custom error rate below 5%
  },
};

export default function () {
  const payload = JSON.stringify({
    grant_type: 'client_credentials',
    client_id: CLIENT_ID,
    client_secret: CLIENT_SECRET,
    scope: 'read write',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const startTime = new Date();
  const response = http.post(`${BASE_URL}/oauth/token`, payload, params);
  const duration = new Date() - startTime;

  // Record custom metrics
  tokenGenerationTime.add(duration);

  // Checks
  const checkResult = check(response, {
    'status is 200': (r) => r.status === 200,
    'has access_token': (r) => r.json('access_token') !== undefined,
    'has refresh_token': (r) => r.json('refresh_token') !== undefined,
    'token_type is Bearer': (r) => r.json('token_type') === 'Bearer',
    'has expires_in': (r) => r.json('expires_in') > 0,
    'response time < 500ms': () => duration < 500,
  });

  // Record errors
  errorRate.add(!checkResult);

  // Brief pause between requests
  sleep(Math.random() * 2 + 1); // Random sleep 1-3 seconds
}

// Teardown function
export function teardown(data) {
  console.log('Load test completed!');
}
