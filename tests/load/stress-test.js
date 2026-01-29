// =============================================================================
// K6 STRESS TEST - FIND BREAKING POINT
// =============================================================================
// Gradually increases load to find system limits
// =============================================================================

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const successRate = new Rate('success');
const responseTime = new Trend('response_time');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const CLIENT_ID = __ENV.CLIENT_ID || 'test_client';
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'test_secret_123456';

// Stress test options - gradually increase load until breaking point
export const options = {
  scenarios: {
    stress_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 20 },    // Ramp up to 20 users
        { duration: '2m', target: 20 },    // Stay at 20
        { duration: '1m', target: 50 },    // Ramp to 50
        { duration: '2m', target: 50 },    // Stay at 50
        { duration: '1m', target: 100 },   // Ramp to 100
        { duration: '2m', target: 100 },   // Stay at 100
        { duration: '1m', target: 150 },   // Ramp to 150
        { duration: '2m', target: 150 },   // Stay at 150
        { duration: '1m', target: 200 },   // Ramp to 200 (stress point)
        { duration: '2m', target: 200 },   // Stay at stress level
        { duration: '2m', target: 0 },     // Ramp down
      ],
      gracefulRampDown: '1m',
    },
  },
  thresholds: {
    // Relaxed thresholds for stress test
    'http_req_duration': ['p(95)<2000', 'p(99)<5000'],
    'http_req_failed': ['rate<0.10'],  // Allow up to 10% errors at peak
    'errors': ['rate<0.15'],
  },
};

export function setup() {
  console.log('Starting stress test...');
  console.log(`Target: ${BASE_URL}`);
  
  // Initial health check
  const health = http.get(`${BASE_URL}/health`);
  if (health.status !== 200) {
    throw new Error('System not healthy before stress test');
  }

  // Get OAuth token
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

  return {
    accessToken: tokenResponse.json('access_token'),
    startTime: new Date(),
  };
}

export default function (data) {
  const accessToken = data.accessToken;

  // Mix of different request types
  const requestType = Math.random();

  let response;
  const startTime = new Date();

  if (requestType < 0.4) {
    // 40% - OAuth token generation (most intensive)
    response = http.post(
      `${BASE_URL}/oauth/token`,
      JSON.stringify({
        grant_type: 'client_credentials',
        client_id: CLIENT_ID,
        client_secret: CLIENT_SECRET,
        scope: 'read write',
      }),
      { headers: { 'Content-Type': 'application/json' } }
    );
  } else if (requestType < 0.7) {
    // 30% - List models
    response = http.get(`${BASE_URL}/v1/models`, {
      headers: { 'Authorization': `Bearer ${accessToken}` },
    });
  } else if (requestType < 0.9) {
    // 20% - Health check (lightest)
    response = http.get(`${BASE_URL}/health`);
  } else {
    // 10% - Metrics
    response = http.get(`${BASE_URL}/metrics`);
  }

  const duration = new Date() - startTime;
  responseTime.add(duration);

  // Check response
  const success = check(response, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time acceptable': () => duration < 3000,
  });

  successRate.add(success);
  errorRate.add(!success);

  // Minimal sleep - we want maximum load
  sleep(0.1 + Math.random() * 0.2); // 100-300ms
}

export function teardown(data) {
  const duration = (new Date() - data.startTime) / 1000;
  console.log('==================================================');
  console.log('Stress Test Completed');
  console.log(`Total Duration: ${duration.toFixed(2)} seconds`);
  console.log('==================================================');
  console.log('Check Grafana dashboards for detailed metrics!');
}
