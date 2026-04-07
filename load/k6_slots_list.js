import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ROOM_COUNT = Number(__ENV.ROOM_COUNT || 20);
const WARM_CACHE = (__ENV.WARM_CACHE || 'false').toLowerCase() === 'true';

export const options = {
  scenarios: {
    slots_list: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 20 },
        { duration: '20s', target: 0 },
      ],
      gracefulRampDown: '5s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<200', 'p(99)<400'],
  },
};

function jsonHeaders(token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

function dummyLogin(role) {
  const res = http.post(`${BASE_URL}/dummyLogin`, JSON.stringify({ role }), {
    headers: jsonHeaders(),
    tags: { endpoint: 'dummyLogin' },
  });

  check(res, {
    [`dummyLogin ${role} status is 200`]: (r) => r.status === 200,
    [`dummyLogin ${role} has token`]: (r) => Boolean(r.json('token')),
  });

  return res.json('token');
}

function isoWeekday(date) {
  const day = date.getUTCDay();
  return day === 0 ? 7 : day;
}

function targetDateString() {
  const tomorrow = new Date();
  tomorrow.setUTCDate(tomorrow.getUTCDate() + 1);
  return tomorrow.toISOString().slice(0, 10);
}

function createRoom(adminToken, index) {
  const res = http.post(
    `${BASE_URL}/rooms/create`,
    JSON.stringify({
      name: `Load Room ${index + 1}`,
      description: 'Room created for k6 load test',
      capacity: 6 + (index % 8),
    }),
    {
      headers: jsonHeaders(adminToken),
      tags: { endpoint: 'rooms_create' },
    },
  );

  check(res, {
    'create room status is 201': (r) => r.status === 201,
    'create room has id': (r) => Boolean(r.json('room.id')),
  });

  return res.json('room.id');
}

function createSchedule(adminToken, roomId, targetDate) {
  const date = new Date(`${targetDate}T00:00:00Z`);
  const res = http.post(
    `${BASE_URL}/rooms/${roomId}/schedule/create`,
    JSON.stringify({
      daysOfWeek: [isoWeekday(date)],
      startTime: '09:00',
      endTime: '18:00',
    }),
    {
      headers: jsonHeaders(adminToken),
      tags: { endpoint: 'schedule_create' },
    },
  );

  check(res, {
    'create schedule status is 201': (r) => r.status === 201,
  });
}

function warmCache(userToken, roomIds, date) {
  for (const roomId of roomIds) {
    const res = http.get(`${BASE_URL}/rooms/${roomId}/slots/list?date=${date}`, {
      headers: jsonHeaders(userToken),
      tags: { endpoint: 'slots_list_warmup' },
    });

    check(res, {
      'warm cache slots status is 200': (r) => r.status === 200,
    });
  }
}

export function setup() {
  const adminToken = dummyLogin('admin');
  const userToken = dummyLogin('user');
  const date = targetDateString();
  const roomIds = [];

  for (let i = 0; i < ROOM_COUNT; i += 1) {
    const roomId = createRoom(adminToken, i);
    createSchedule(adminToken, roomId, date);
    roomIds.push(roomId);
  }

  if (WARM_CACHE) {
    warmCache(userToken, roomIds, date);
  }

  return { userToken, roomIds, date };
}

export default function (data) {
  const roomId = data.roomIds[Math.floor(Math.random() * data.roomIds.length)];
  const res = http.get(`${BASE_URL}/rooms/${roomId}/slots/list?date=${data.date}`, {
    headers: jsonHeaders(data.userToken),
    tags: { endpoint: 'slots_list' },
  });

  check(res, {
    'slots list status is 200': (r) => r.status === 200,
    'slots list returns array': (r) => Array.isArray(r.json('slots')),
  });

  sleep(0.5);
}
