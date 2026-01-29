// =============================================================================
// K6 LOAD TEST - API ENDPOINTS
// =============================================================================
// Tests multiple API endpoints under realistic load
// =============================================================================

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration');
const requestsTotal = new Counter('requests_total');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const CLIENT_ID = __ENV.CLIENT_ID || 'test_client';
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'test_secret_123456';
const ADMIN_API_KEY = __ENV.ADMIN_API_KEY || 'admin_dev_key_12345678901234567890123456789012';

// Load test options
export const options = {
  scenarios: {
    // Realistic user behavior simulation
    realistic_load: {
      executor: 'ramping-arrival-rate',
      startRate: 5,    // Start with 5 requests per second
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 100,
      stages: [
        { duration: '1m', target: 10 },   // Ramp up to 10 RPS
        { duration: '3m', target: 10 },   // Stay at 10 RPS
        { duration: '1m', target: 30 },   // Spike to 30 RPS
        { duration: '2m', target: 30 },   // Stay at peak
        { duration: '1m', target: 5 },    // Ramp down
      ],
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<1000', 'p(99)<2000'],
    'http_req_failed': ['rate<0.02'],
    'errors': ['rate<0.05'],
  },
};

// Setup function - runs once at the beginning
export function setup() {
  console.log('Setting up load test...');
  
  // Get an OAuth token for use in tests
  const tokenResponse = http.post(
    `${BASE_URL}/oauth/token`,
    JSON.stringify({
      grant_type: 'client_credentials',
      client_id: CLIENT_ID,
      client_secret: CLIENT_SECRET,
      scope: 'read write',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  if (tokenResponse.status !== 200) {
    throw new Error('Failed to get OAuth token in setup');
  }

  return {
    accessToken: tokenResponse.json('access_token'),
  };
}

export default function (data) {
  const accessToken = data.accessToken;
  requestsTotal.add(1);

  // Scenario: Random user workflow
  const scenario = Math.random();

  if (scenario < 0.3) {
    // 30% - Health check
    testHealthCheck();
  } else if (scenario < 0.6) {
    // 30% - List models
    testListModels(accessToken);
  } else if (scenario < 0.8) {
    // 20% - Get single model
    testGetModel(accessToken);
  } else if (scenario < 0.9) {
    // 10% - Admin endpoints
    testAdminEndpoints();
  } else {
    // 10% - Metrics endpoint
    testMetricsEndpoint();
  }

  sleep(Math.random() * 3 + 1); // 1-4 seconds between requests
}

function testHealthCheck() {
  group('Health Check', function () {
    const startTime = new Date();
    const response = http.get(`${BASE_URL}/health`);
    const duration = new Date() - startTime;

    requestDuration.add(duration, { endpoint: 'health' });

    const result = check(response, {
      'health: status is 200': (r) => r.status === 200,
      'health: response time < 100ms': () => duration < 100,
      'health: status is healthy': (r) => r.json('status') === 'healthy',
    });

    errorRate.add(!result);
  });
}

function testListModels(accessToken) {
  group('List Models', function () {
    const startTime = new Date();
    const response = http.get(`${BASE_URL}/v1/models`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });
    const duration = new Date() - startTime;

    requestDuration.add(duration, { endpoint: 'list_models' });

    const result = check(response, {
      'models: status is 200': (r) => r.status === 200,
      'models: has data': (r) => r.json('data') !== undefined,
      'models: response time < 500ms': () => duration < 500,
    });

    errorRate.add(!result);
  });
}

function testGetModel(accessToken) {
  group('Get Model', function () {
    const models = [
      'claude-3-opus-20240229',
      'claude-3-sonnet-20240229',
      'claude-3-haiku-20240307',
    ];
    const model = models[Math.floor(Math.random() * models.length)];

    const startTime = new Date();
    const response = http.get(`${BASE_URL}/v1/models/${model}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });
    const duration = new Date() - startTime;

    requestDuration.add(duration, { endpoint: 'get_model' });

    const result = check(response, {
      'get_model: status is 200': (r) => r.status === 200,
      'get_model: has id': (r) => r.json('id') !== undefined,
      'get_model: response time < 500ms': () => duration < 500,
    });

    errorRate.add(!result);
  });
}

function testAdminEndpoints() {
  group('Admin Endpoints', function () {
    const endpoints = [
      '/admin/clients',
      '/admin/cache/stats',
      '/admin/providers/status',
    ];
    const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];

    const startTime = new Date();
    const response = http.get(`${BASE_URL}${endpoint}`, {
      headers: {
        'X-Admin-API-Key': ADMIN_API_KEY,
      },
    });
    const duration = new Date() - startTime;

    requestDuration.add(duration, { endpoint: 'admin' });

    const result = check(response, {
      'admin: status is 200': (r) => r.status === 200,
      'admin: has valid JSON': (r) => r.json() !== undefined,
      'admin: response time < 1000ms': () => duration < 1000,
    });

    errorRate.add(!result);
  });
}

function testMetricsEndpoint() {
  group('Metrics Endpoint', function () {
    const startTime = new Date();
    const response = http.get(`${BASE_URL}/metrics`);
    const duration = new Date() - startTime;

    requestDuration.add(duration, { endpoint: 'metrics' });

    const result = check(response, {
      'metrics: status is 200': (r) => r.status === 200,
      'metrics: contains prometheus data': (r) => r.body.includes('llm_proxy_'),
      'metrics: response time < 500ms': () => duration < 500,
    });

    errorRate.add(!result);
  });
}

// Teardown function
export function teardown(data) {
  console.log('Load test completed!');
  console.log(`Access token used: ${data.accessToken.substring(0, 20)}...`);
}
