import http from 'k6/http';

const BASE_URL = 'http://localhost:8080';

export function setup() {
  console.log('Setting up test data...');

  const teamPayload = JSON.stringify({
    team_name: "loadtest-team",
    members: [
      { user_id: "loadtest-1", username: "LoadTestUser1", is_active: true },
      { user_id: "loadtest-2", username: "LoadTestUser2", is_active: true },
      { user_id: "loadtest-3", username: "LoadTestUser3", is_active: true },
      { user_id: "loadtest-4", username: "LoadTestUser4", is_active: true },
      { user_id: "loadtest-5", username: "LoadTestUser5", is_active: true },
    ]
  });

  const teamResponse = http.post(`${BASE_URL}/team/add`, teamPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  for (let i = 1; i <= 10; i++) {
    const prPayload = JSON.stringify({
      pull_request_id: `loadtest-pr-${i}`,
      pull_request_name: `Load Test PR ${i}`,
      author_id: `loadtest-${(i % 5) + 1}`
    });

    http.post(`${BASE_URL}/pullRequest/create`, prPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
  }

  console.log('Test data setup completed');
  return { teamId: "loadtest-team" };
}

export function teardown(data) {
  console.log('Cleaning up test data...');
}