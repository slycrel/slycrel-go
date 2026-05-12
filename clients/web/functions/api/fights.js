import { jsonResponse, errorResponse } from '../_lib/json.js';

export async function onRequestGet({ env }) {
  const row = await env.DB.prepare(
    "SELECT value FROM singletons WHERE key = 'gladiator_fights'"
  ).first();
  return jsonResponse(row ? JSON.parse(row.value) : []);
}

export async function onRequestPut({ request, env }) {
  const fights = await request.json().catch(() => null);
  if (!Array.isArray(fights)) return errorResponse(400, 'array body required');
  await env.DB.prepare(
    "INSERT INTO singletons (key, value) VALUES ('gladiator_fights', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value"
  ).bind(JSON.stringify(fights)).run();
  return jsonResponse({ ok: true });
}
