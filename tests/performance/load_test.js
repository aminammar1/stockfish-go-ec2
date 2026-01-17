import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 5,
  duration: '10s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  let healthRes = http.get(`${BASE_URL}/api/v1/health`);
  check(healthRes, {
    'health status is 200': (r) => r.status === 200,
  });

  let analyzeRes = http.post(
    `${BASE_URL}/api/v1/analyze`,
    JSON.stringify({ fen: 'rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  check(analyzeRes, {
    'analyze status is 200': (r) => r.status === 200,
  });

  sleep(1);
}
