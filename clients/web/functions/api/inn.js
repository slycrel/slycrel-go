import { jsonResponse, errorResponse } from '../_lib/json.js';

export async function onRequestGet({ env }) {
  const row = await env.DB.prepare(
    "SELECT value FROM singletons WHERE key = 'inn'"
  ).first();
  return jsonResponse(row ? JSON.parse(row.value) : null);
}

export async function onRequestPut({ request, env }) {
  const inn = await request.json().catch(() => null);
  if (!inn) return errorResponse(400, 'inn body required');
  await env.DB.prepare(
    "INSERT INTO singletons (key, value) VALUES ('inn', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value"
  ).bind(JSON.stringify(inn)).run();
  return jsonResponse({ ok: true });
}
