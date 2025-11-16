import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '20s', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '20s', target: 10 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = 'http://localhost:8080';

export default function () {
  const statsResponse = http.get(`${BASE_URL}/stats/get`);
  check(statsResponse, {
    'stats status is 200': (r) => r.status === 200,
    'stats response time < 500ms': (r) => r.timings.duration < 500,
  });

  const reviewResponse = http.get(`${BASE_URL}/users/getReview?user_id=test-1`);
  check(reviewResponse, {
    'review status is 200': (r) => r.status === 200,
    'review response time < 300ms': (r) => r.timings.duration < 300,
  });

  const teamResponse = http.get(`${BASE_URL}/team/get?team_name=test-team`);
  check(teamResponse, {
    'team status is 200': (r) => r.status === 200,
    'team response time < 400ms': (r) => r.timings.duration < 400,
  });

  sleep(1);
}