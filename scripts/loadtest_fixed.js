import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:8080';

export function setup() {
  console.log('Setting up test data for load testing...');

  const teamPayload = JSON.stringify({
    team_name: "loadtest-team",
    members: [
      { user_id: "loadtest-1", username: "LoadTest1", is_active: true },
      { user_id: "loadtest-2", username: "LoadTest2", is_active: true },
      { user_id: "loadtest-3", username: "LoadTest3", is_active: true },
      { user_id: "loadtest-4", username: "LoadTest4", is_active: true },
      { user_id: "loadtest-5", username: "LoadTest5", is_active: true },
    ]
  });

  const teamRes = http.post(`${BASE_URL}/team/add`, teamPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (teamRes.status !== 201) {
    console.log(`Team setup failed: ${teamRes.status} - ${teamRes.body}`);
  } else {
    console.log('Team setup successful');
  }

  let prCount = 0;
  for (let i = 1; i <= 10; i++) {
    const prPayload = JSON.stringify({
      pull_request_id: `loadtest-pr-${i}`,
      pull_request_name: `Load Test PR ${i}`,
      author_id: `loadtest-${(i % 5) + 1}`
    });

    const prRes = http.post(`${BASE_URL}/pullRequest/create`, prPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (prRes.status === 201) {
      prCount++;
    }
  }

  console.log(`Created ${prCount} PRs successfully`);

  return { 
    teamName: "loadtest-team",
    userIds: ["loadtest-1", "loadtest-2", "loadtest-3", "loadtest-4", "loadtest-5"]
  };
}

export const options = {
  stages: [
    { duration: '30s', target: 5 },
    { duration: '1m', target: 10 },
    { duration: '1m', target: 20 },
    { duration: '30s', target: 10 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<100'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function (data) {
  const randomUser = data.userIds[Math.floor(Math.random() * data.userIds.length)];

  const statsRes = http.get(`${BASE_URL}/stats/get`);
  check(statsRes, {
    'stats status 200': (r) => r.status === 200,
    'stats time < 50ms': (r) => r.timings.duration < 50,
  });

  const reviewRes = http.get(`${BASE_URL}/users/getReview?user_id=${randomUser}`);
  check(reviewRes, {
    'review status 200': (r) => r.status === 200,
    'review time < 100ms': (r) => r.timings.duration < 100,
  });

  const teamRes = http.get(`${BASE_URL}/team/get?team_name=${data.teamName}`);
  check(teamRes, {
    'team status 200': (r) => r.status === 200,
    'team time < 100ms': (r) => r.timings.duration < 100,
  });

  sleep(0.5);
}