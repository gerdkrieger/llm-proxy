// =============================================================================
// K6 SPIKE TEST - SUDDEN TRAFFIC SURGE
// =============================================================================
// Tests system behavior under sudden traffic spikes
// =============================================================================

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const spikeRecoveryTime = new Trend('spike_recovery_time');
const requestsInSpike = new Counter('requests_in_spike');

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const CLIENT_ID = __ENV.CLIENT_ID || 'test_client';
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'test_secret_123456';

// Spike test options
export const options = {
  scenarios: {
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 10 },   // Normal load
        { duration: '1m', target: 10 },    // Stay normal
        { duration: '10s', target: 200 },  // SPIKE! 10s -> 200 users
        { duration: '1m', target: 200 },   // Sustain spike
        { duration: '10s', target: 10 },   // Drop back to normal
        { duration: '1m', target: 10 },    // Recovery period
        { duration: '30s', target: 0 },    // Ramp down
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<3000'],  // More lenient during spike
    'http_req_failed': ['rate<0.15'],     // Allow up to 15% errors during spike
    'errors': ['rate<0.20'],
  },
};

export function setup() {
  console.log('Preparing for spike test...');
  
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
  };
}

export default function (data) {
  const accessToken = data.accessToken;

  // Track if we're in spike window (rough approximation)
  const isInSpike = __VU > 50; // More than 50 VUs = spike period
  
  if (isInSpike) {
    requestsInSpike.add(1);
  }

  // Random request
  const actions = [
    // OAuth tokens
    () => {
      const response = http.post(
        `${BASE_URL}/oauth/token`,
        JSON.stringify({
          grant_type: 'client_credentials',
          client_id: CLIENT_ID,
          client_secret: CLIENT_SECRET,
          scope: 'read write',
        }),
        { headers: { 'Content-Type': 'application/json' } }
      );
      return response;
    },
    // List models
    () => {
      return http.get(`${BASE_URL}/v1/models`, {
        headers: { 'Authorization': `Bearer ${accessToken}` },
      });
    },
    // Health check
    () => {
      return http.get(`${BASE_URL}/health`);
    },
  ];

  const action = actions[Math.floor(Math.random() * actions.length)];
  const startTime = new Date();
  const response = action();
  const duration = new Date() - startTime;

  if (isInSpike) {
    spikeRecoveryTime.add(duration);
  }

  const result = check(response, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
    'response within reasonable time': () => duration < 5000,
  });

  errorRate.add(!result);

  // Very short sleep during spike
  sleep(isInSpike ? 0.05 : 0.5);
}

export function teardown(data) {
  console.log('==================================================');
  console.log('Spike Test Completed');
  console.log('Check metrics to see how system handled spike!');
  console.log('==================================================');
}
