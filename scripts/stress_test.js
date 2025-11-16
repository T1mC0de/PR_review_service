import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://localhost:8080';

export function setup() {
  console.log('Setting up test data for stress testing...');

  const teamPayload = JSON.stringify({
    team_name: "stress-team",
    members: [
      { user_id: "stress-user-1", username: "StressUser1", is_active: true },
      { user_id: "stress-user-2", username: "StressUser2", is_active: true },
      { user_id: "stress-user-3", username: "StressUser3", is_active: true },
      { user_id: "stress-user-4", username: "StressUser4", is_active: true },
      { user_id: "stress-user-5", username: "StressUser5", is_active: true },
    ]
  });

  const teamRes = http.post(`${BASE_URL}/team/add`, teamPayload, {
    headers: { 'Content-Type': 'application/json' },
  });
  
  if (teamRes.status !== 201) {
    console.log(`Team setup failed: ${teamRes.status} - ${teamRes.body}`);
    throw new Error('Team setup failed');
  } else {
    console.log('Team setup successful');
  }

  const prMeta = [];
  let prCount = 0;
  for (let i = 1; i <= 10; i++) {
    const authorId = `stress-user-${(i % 5) + 1}`;
    const prId = `stress-pr-${i}`;
    const prPayload = JSON.stringify({
      pull_request_id: prId,
      pull_request_name: `Stress Test PR ${i}`,
      author_id: authorId
    });
    
    const prRes = http.post(`${BASE_URL}/pullRequest/create`, prPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (prRes.status === 201) {
      prMeta.push({ id: prId, author: authorId });
      prCount++;
    } else {
      console.log(`PR creation failed: ${prRes.status} - ${prRes.body}`);
    }
  }
  
  console.log(`Created ${prCount} PRs successfully`);

  return {
    teamName: 'stress-team',
    userIds: ['stress-user-1', 'stress-user-2', 'stress-user-3', 'stress-user-4', 'stress-user-5'],
    prMeta
  };
}

export const options = {
  stages: [
    { duration: '2m', target: 5 },
    { duration: '3m', target: 8 },
    { duration: '2m', target: 12 },
    { duration: '2m', target: 5 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<350'],
    http_req_failed: ['rate<0.005'],
  },
};

export default function (data) {
  if (!data || !data.userIds) {
    console.error('Setup data not available!');
    return;
  }
  
  const randomUser = data.userIds[Math.floor(Math.random() * data.userIds.length)];
  const op = Math.random();
  
  let method = 'GET';
  let url, body;
  const headers = { 'Content-Type': 'application/json' };

  if (op < 0.70) {
    const readEndpoints = [
      `${BASE_URL}/stats/get`,
      `${BASE_URL}/team/get?team_name=${data.teamName}`,
      `${BASE_URL}/users/getReview?user_id=${randomUser}`,
    ];
    url = readEndpoints[Math.floor(Math.random() * readEndpoints.length)];
    method = 'GET';
  } else if (op < 0.85) {
    method = 'POST';
    const prId = `stress-dyn-${Date.now()}-${Math.random().toString(36).substr(2, 5)}`;
    url = `${BASE_URL}/pullRequest/create`;
    body = JSON.stringify({ 
      pull_request_id: prId, 
      pull_request_name: `Dynamic PR ${prId}`, 
      author_id: randomUser 
    });
  } else if (op < 0.95) {
    method = 'POST';
    url = `${BASE_URL}/users/setIsActive`;
    body = JSON.stringify({ 
      user_id: randomUser, 
      is_active: Math.random() < 0.3
    });
  } else {
    method = 'POST';
    if (data.prMeta && data.prMeta.length > 0) {
      const targetPr = data.prMeta[Math.floor(Math.random() * data.prMeta.length)];
      url = `${BASE_URL}/pullRequest/merge`;
      body = JSON.stringify({ pull_request_id: targetPr.id });
    } else {
      url = `${BASE_URL}/stats/get`;
      method = 'GET';
    }
  }

  const params = {
    headers: headers,
    timeout: '5s'
  };

  let res;
  try {
    res = method === 'GET' ? http.get(url, params) : http.post(url, body, params);
  } catch (error) {
    console.log(`Request failed: ${error.message} for ${url}`);
    return;
  }

  check(res, {
    'status 2xx': (r) => r.status >= 200 && r.status < 300,
    'response time <350ms': (r) => r.timings.duration < 350,
    'no server errors': (r) => r.status < 500,
  });

  if (res.status >= 500) {
    console.log(`SERVER ERROR ${res.status}: ${method} ${url}`);
  } else if (res.status >= 400 && res.status !== 404) {
    console.log(`CLIENT ERROR ${res.status}: ${method} ${url} -> ${res.body.substring(0, 80)}`);
  } else if (res.timings.duration > 350) {
    console.log(`SLOW RESPONSE ${res.timings.duration}ms: ${method} ${url}`);
  }

  const baseSleep = 0.5;
  const adaptiveSleep = res.status >= 400 ? baseSleep * 2 : baseSleep;
  sleep(adaptiveSleep);
}