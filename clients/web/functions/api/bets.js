import { jsonResponse, errorResponse } from '../_lib/json.js';

export async function onRequestGet({ env }) {
  const row = await env.DB.prepare(
    "SELECT value FROM singletons WHERE key = 'gladiator_bets'"
  ).first();
  return jsonResponse(row ? JSON.parse(row.value) : []);
}

export async function onRequestPut({ request, env }) {
  const bets = await request.json().catch(() => null);
  if (!Array.isArray(bets)) return errorResponse(400, 'array body required');
  await env.DB.prepare(
    "INSERT INTO singletons (key, value) VALUES ('gladiator_bets', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value"
  ).bind(JSON.stringify(bets)).run();
  return jsonResponse({ ok: true });
}
